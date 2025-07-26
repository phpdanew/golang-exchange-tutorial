package auth

import (
	"context"
	"database/sql"
	"testing"

	"crypto-exchange/internal/config"
	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"golang.org/x/crypto/bcrypt"
)

// MockUserModel 模拟用户模型
type MockUserModel struct {
	mock.Mock
}

func (m *MockUserModel) Insert(ctx context.Context, data *model.User) (sql.Result, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(sql.Result), args.Error(1)
}

func (m *MockUserModel) FindOne(ctx context.Context, id uint64) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserModel) FindOneByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserModel) Update(ctx context.Context, data *model.User) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockUserModel) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockResult 模拟SQL结果
type MockResult struct {
	mock.Mock
}

func (m MockResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m MockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestRegisterLogic_Register(t *testing.T) {
	tests := []struct {
		name    string
		req     *types.RegisterRequest
		setup   func(*MockUserModel, *MockResult)
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功注册用户",
			req: &types.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Nickname: "testuser",
			},
			setup: func(mockUser *MockUserModel, mockResult *MockResult) {
				// 模拟用户不存在
				mockUser.On("FindOneByEmail", mock.Anything, "test@example.com").Return(nil, model.ErrNotFound)
				
				// 模拟插入成功
				mockResult.On("LastInsertId").Return(int64(1), nil)
				mockUser.On("Insert", mock.Anything, mock.AnythingOfType("*model.User")).Return(mockResult, nil)
			},
			wantErr: false,
		},
		{
			name: "邮箱格式无效",
			req: &types.RegisterRequest{
				Email:    "invalid-email",
				Password: "password123",
				Nickname: "testuser",
			},
			setup:   func(*MockUserModel, *MockResult) {},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "密码太短",
			req: &types.RegisterRequest{
				Email:    "test@example.com",
				Password: "123",
				Nickname: "testuser",
			},
			setup:   func(*MockUserModel, *MockResult) {},
			wantErr: true,
			errMsg:  "password must be at least 6 characters long",
		},
		{
			name: "密码缺少字母",
			req: &types.RegisterRequest{
				Email:    "test@example.com",
				Password: "123456",
				Nickname: "testuser",
			},
			setup:   func(*MockUserModel, *MockResult) {},
			wantErr: true,
			errMsg:  "password must contain at least one letter and one number",
		},
		{
			name: "密码缺少数字",
			req: &types.RegisterRequest{
				Email:    "test@example.com",
				Password: "password",
				Nickname: "testuser",
			},
			setup:   func(*MockUserModel, *MockResult) {},
			wantErr: true,
			errMsg:  "password must contain at least one letter and one number",
		},
		{
			name: "用户已存在",
			req: &types.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
				Nickname: "testuser",
			},
			setup: func(mockUser *MockUserModel, mockResult *MockResult) {
				// 模拟用户已存在
				existingUser := &model.User{
					ID:    1,
					Email: "existing@example.com",
				}
				mockUser.On("FindOneByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			wantErr: true,
			errMsg:  "user already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			mockUser := &MockUserModel{}
			mockResult := &MockResult{}
			
			// 设置模拟行为
			tt.setup(mockUser, mockResult)

			// 创建服务上下文
			svcCtx := &svc.ServiceContext{
				Config: config.Config{},
				UserModel: mockUser,
				RedisClient: &redis.Redis{},
			}

			// 创建逻辑实例
			logic := NewRegisterLogic(context.Background(), svcCtx)

			// 执行测试
			resp, err := logic.Register(tt.req)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.Email, resp.Email)
				assert.Equal(t, tt.req.Nickname, resp.Nickname)
				assert.Equal(t, int64(1), resp.Status)
				assert.Greater(t, resp.ID, uint64(0))
			}

			// 验证模拟调用
			mockUser.AssertExpectations(t)
			if !tt.wantErr && tt.name == "成功注册用户" {
				mockResult.AssertExpectations(t)
			}
		})
	}
}

func TestRegisterLogic_validateEmail(t *testing.T) {
	logic := &RegisterLogic{}

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"有效邮箱", "test@example.com", false},
		{"有效邮箱带数字", "user123@domain.org", false},
		{"有效邮箱带特殊字符", "user.name+tag@example.co.uk", false},
		{"空邮箱", "", true},
		{"无@符号", "testexample.com", true},
		{"无域名", "test@", true},
		{"无用户名", "@example.com", true},
		{"无顶级域名", "test@example", true},
		{"多个@符号", "test@@example.com", true},
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

func TestRegisterLogic_validatePassword(t *testing.T) {
	logic := &RegisterLogic{}

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{"有效密码", "password123", false, ""},
		{"有效密码大写", "Password123", false, ""},
		{"有效密码特殊字符", "pass@123", false, ""},
		{"密码太短", "12345", true, "password must be at least 6 characters long"},
		{"只有字母", "password", true, "password must contain at least one letter and one number"},
		{"只有数字", "123456", true, "password must contain at least one letter and one number"},
		{"只有特殊字符", "!@#$%^", true, "password must contain at least one letter and one number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logic.validatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisterLogic_hashPassword(t *testing.T) {
	logic := &RegisterLogic{}

	password := "testpassword123"
	hashedPassword, err := logic.hashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	// 验证哈希密码可以被验证
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err)

	// 验证错误密码无法通过验证
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("wrongpassword"))
	assert.Error(t, err)
}