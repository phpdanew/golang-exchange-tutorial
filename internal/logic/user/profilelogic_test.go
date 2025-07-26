package user

import (
	"context"
	"database/sql"
	"testing"

	"crypto-exchange/internal/config"
	"crypto-exchange/internal/svc"
	"crypto-exchange/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
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

func TestProfileLogic_Profile(t *testing.T) {
	tests := []struct {
		name      string
		userID    interface{} // 模拟context中的用户ID
		setupUser func(*MockUserModel)
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "成功获取用户信息",
			userID: float64(1), // JWT中的数值通常是float64
			setupUser: func(mockUser *MockUserModel) {
				user := &model.User{
					ID:       1,
					Email:    "test@example.com",
					Nickname: "testuser",
					Status:   1,
				}
				mockUser.On("FindOne", mock.Anything, uint64(1)).Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:   "用户不存在",
			userID: float64(999),
			setupUser: func(mockUser *MockUserModel) {
				mockUser.On("FindOne", mock.Anything, uint64(999)).Return(nil, model.ErrNotFound)
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name:   "用户被禁用",
			userID: float64(2),
			setupUser: func(mockUser *MockUserModel) {
				user := &model.User{
					ID:       2,
					Email:    "disabled@example.com",
					Nickname: "disableduser",
					Status:   2, // 禁用状态
				}
				mockUser.On("FindOne", mock.Anything, uint64(2)).Return(user, nil)
			},
			wantErr: true,
			errMsg:  "user account is disabled",
		},
		{
			name:      "上下文中没有用户ID",
			userID:    nil,
			setupUser: func(mockUser *MockUserModel) {
				// 不会调用用户查询
			},
			wantErr: true,
			errMsg:  "unauthorized",
		},
		{
			name:   "用户ID类型为uint64",
			userID: uint64(3),
			setupUser: func(mockUser *MockUserModel) {
				user := &model.User{
					ID:       3,
					Email:    "test3@example.com",
					Nickname: "testuser3",
					Status:   1,
				}
				mockUser.On("FindOne", mock.Anything, uint64(3)).Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:   "用户ID类型为int64",
			userID: int64(4),
			setupUser: func(mockUser *MockUserModel) {
				user := &model.User{
					ID:       4,
					Email:    "test4@example.com",
					Nickname: "testuser4",
					Status:   1,
				}
				mockUser.On("FindOne", mock.Anything, uint64(4)).Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:   "用户ID类型为int",
			userID: int(5),
			setupUser: func(mockUser *MockUserModel) {
				user := &model.User{
					ID:       5,
					Email:    "test5@example.com",
					Nickname: "testuser5",
					Status:   1,
				}
				mockUser.On("FindOne", mock.Anything, uint64(5)).Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:      "无效的用户ID类型",
			userID:    "invalid",
			setupUser: func(mockUser *MockUserModel) {
				// 不会调用用户查询
			},
			wantErr: true,
			errMsg:  "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟对象
			mockUser := &MockUserModel{}

			// 设置模拟行为
			tt.setupUser(mockUser)

			// 创建带有用户ID的上下文
			ctx := context.Background()
			if tt.userID != nil {
				ctx = context.WithValue(ctx, "userId", tt.userID)
			}

			// 创建服务上下文
			svcCtx := &svc.ServiceContext{
				Config:      config.Config{},
				UserModel:   mockUser,
				RedisClient: &redis.Redis{},
			}

			// 创建逻辑实例
			logic := NewProfileLogic(ctx, svcCtx)

			// 执行测试
			resp, err := logic.Profile()

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Greater(t, resp.ID, uint64(0))
				assert.NotEmpty(t, resp.Email)
				assert.NotEmpty(t, resp.Nickname)
				assert.Equal(t, int64(1), resp.Status)
			}

			// 验证模拟调用
			mockUser.AssertExpectations(t)
		})
	}
}

func TestProfileLogic_getUserIDFromContext(t *testing.T) {
	tests := []struct {
		name    string
		userID  interface{}
		want    uint64
		wantErr bool
	}{
		{
			name:    "float64类型",
			userID:  float64(123),
			want:    123,
			wantErr: false,
		},
		{
			name:    "uint64类型",
			userID:  uint64(456),
			want:    456,
			wantErr: false,
		},
		{
			name:    "int64类型",
			userID:  int64(789),
			want:    789,
			wantErr: false,
		},
		{
			name:    "int类型",
			userID:  int(101112),
			want:    101112,
			wantErr: false,
		},
		{
			name:    "nil值",
			userID:  nil,
			want:    0,
			wantErr: true,
		},
		{
			name:    "string类型（无效）",
			userID:  "123",
			want:    0,
			wantErr: true,
		},
		{
			name:    "bool类型（无效）",
			userID:  true,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建带有用户ID的上下文
			ctx := context.Background()
			if tt.userID != nil {
				ctx = context.WithValue(ctx, "userId", tt.userID)
			}

			// 创建逻辑实例
			logic := &ProfileLogic{
				Logger: logx.WithContext(ctx),
				ctx:    ctx,
			}

			// 执行测试
			got, err := logic.getUserIDFromContext()

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, uint64(0), got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}