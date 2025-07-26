package auth

import (
	"context"
	"testing"

	"crypto-exchange/internal/config"
	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginLogic_Login_BasicCases(t *testing.T) {
	// 创建测试用的哈希密码
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		req       *types.LoginRequest
		setupUser func(*MockUserModel)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "用户不存在",
			req: &types.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setupUser: func(mockUser *MockUserModel) {
				mockUser.On("FindOneByEmail", mock.Anything, "nonexistent@example.com").Return(nil, model.ErrNotFound)
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name: "用户被禁用",
			req: &types.LoginRequest{
				Email:    "disabled@example.com",
				Password: "password123",
			},
			setupUser: func(mockUser *MockUserModel) {
				user := &model.User{
					ID:       1,
					Email:    "disabled@example.com",
					Password: string(hashedPassword),
					Nickname: "disableduser",
					Status:   2, // 禁用状态
				}
				mockUser.On("FindOneByEmail", mock.Anything, "disabled@example.com").Return(user, nil)
			},
			wantErr: true,
			errMsg:  "user account is disabled",
		},
		{
			name: "邮箱为空",
			req: &types.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			setupUser: func(mockUser *MockUserModel) {
				// 不会调用用户查询
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			mockUser := &MockUserModel{}

			// 设置模拟行为
			tt.setupUser(mockUser)

			// 创建服务上下文
			svcCtx := &svc.ServiceContext{
				Config: config.Config{
					Auth: struct {
						AccessSecret string
						AccessExpire int64
					}{
						AccessSecret: "test-secret",
						AccessExpire: 3600,
					},
				},
				UserModel:   mockUser,
				RedisClient: &redis.Redis{}, // 使用真实的Redis客户端，但在测试中不会实际调用
			}

			// 创建逻辑实例
			logic := NewLoginLogic(context.Background(), svcCtx)

			// 执行测试
			resp, err := logic.Login(tt.req)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)
				assert.Equal(t, tt.req.Email, resp.User.Email)
				assert.Greater(t, resp.User.ID, uint64(0))
			}

			// 验证模拟调用
			mockUser.AssertExpectations(t)
		})
	}
}

func TestLoginLogic_verifyPassword(t *testing.T) {
	logic := &LoginLogic{}

	// 创建测试密码和哈希
	password := "testpassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		plainPassword  string
		hashedPassword string
		wantErr        bool
	}{
		{
			name:           "正确密码",
			plainPassword:  password,
			hashedPassword: string(hashedPassword),
			wantErr:        false,
		},
		{
			name:           "错误密码",
			plainPassword:  "wrongpassword",
			hashedPassword: string(hashedPassword),
			wantErr:        true,
		},
		{
			name:           "空密码",
			plainPassword:  "",
			hashedPassword: string(hashedPassword),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logic.verifyPassword(tt.plainPassword, tt.hashedPassword)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, model.ErrInvalidPassword, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoginLogic_generateJWTToken(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		Config: config.Config{
			Auth: struct {
				AccessSecret string
				AccessExpire int64
			}{
				AccessSecret: "test-secret-key",
				AccessExpire: 3600,
			},
		},
	}

	logic := &LoginLogic{
		svcCtx: svcCtx,
	}

	user := &model.User{
		ID:       123,
		Email:    "test@example.com",
		Nickname: "testuser",
		Status:   1,
	}

	token, err := logic.generateJWTToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证token可以被解析
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, float64(123), claims["userId"])
	assert.Equal(t, "test@example.com", claims["email"])

	// 验证过期时间
	exp := claims["exp"].(float64)
	iat := claims["iat"].(float64)
	assert.Equal(t, float64(3600), exp-iat)
}

func TestLoginLogic_validateEmail(t *testing.T) {
	logic := &LoginLogic{}

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"有效邮箱", "test@example.com", false},
		{"空邮箱", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logic.validateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, model.ErrInvalidEmail, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}