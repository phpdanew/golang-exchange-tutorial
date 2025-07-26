package asset

import (
	"context"
	"testing"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/logx"
)

// 使用 depositlogic_test.go 中的 MockBalanceModel

func TestGetBalancesLogic_GetBalances(t *testing.T) {
	// 禁用日志输出以保持测试输出清洁
	logx.DisableStat()

	tests := []struct {
		name           string
		userID         uint64
		mockBalances   []*model.Balance
		mockError      error
		expectedResult *types.BalanceResponse
		expectedError  error
		setupContext   func() context.Context
	}{
		{
			name:   "成功查询用户余额",
			userID: 1,
			mockBalances: []*model.Balance{
				{
					ID:        1,
					UserID:    1,
					Currency:  "BTC",
					Available: "1.50000000",
					Frozen:    "0.25000000",
					UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				{
					ID:        2,
					UserID:    1,
					Currency:  "USDT",
					Available: "10000.00000000",
					Frozen:    "500.00000000",
					UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			mockError: nil,
			expectedResult: &types.BalanceResponse{
				Balances: []types.Balance{
					{
						Currency:  "BTC",
						Available: "1.50000000",
						Frozen:    "0.25000000",
						UpdatedAt: "2024-01-01 12:00:00",
					},
					{
						Currency:  "USDT",
						Available: "10000.00000000",
						Frozen:    "500.00000000",
						UpdatedAt: "2024-01-01 12:00:00",
					},
				},
			},
			expectedError: nil,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(1))
			},
		},
		{
			name:           "用户无余额记录",
			userID:         2,
			mockBalances:   []*model.Balance{},
			mockError:      nil,
			expectedResult: &types.BalanceResponse{Balances: []types.Balance{}},
			expectedError:  nil,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(2))
			},
		},
		{
			name:           "数据库查询错误",
			userID:         3,
			mockBalances:   nil,
			mockError:      assert.AnError,
			expectedResult: nil,
			expectedError:  model.ErrInternalServer,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(3))
			},
		},
		{
			name:           "JWT上下文中无用户ID",
			userID:         0,
			mockBalances:   nil,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  model.ErrUnauthorized,
			setupContext: func() context.Context {
				return context.Background() // 没有userId
			},
		},
		{
			name:   "包含无效余额格式的记录",
			userID: 4,
			mockBalances: []*model.Balance{
				{
					ID:        1,
					UserID:    4,
					Currency:  "BTC",
					Available: "1.50000000",
					Frozen:    "0.25000000",
					UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				{
					ID:        2,
					UserID:    4,
					Currency:  "ETH",
					Available: "invalid_amount", // 无效格式
					Frozen:    "0.00000000",
					UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				{
					ID:        3,
					UserID:    4,
					Currency:  "USDT",
					Available: "1000.00000000",
					Frozen:    "0.00000000",
					UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			mockError: nil,
			expectedResult: &types.BalanceResponse{
				Balances: []types.Balance{
					{
						Currency:  "BTC",
						Available: "1.50000000",
						Frozen:    "0.25000000",
						UpdatedAt: "2024-01-01 12:00:00",
					},
					{
						Currency:  "USDT",
						Available: "1000.00000000",
						Frozen:    "0.00000000",
						UpdatedAt: "2024-01-01 12:00:00",
					},
				},
			},
			expectedError: nil,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(4))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟的BalanceModel
			mockBalanceModel := new(MockBalanceModel)
			
			// 设置模拟期望（只有在需要调用数据库时才设置）
			if tt.userID > 0 {
				mockBalanceModel.On("FindByUserID", mock.Anything, tt.userID).Return(tt.mockBalances, tt.mockError)
			}

			// 创建服务上下文
			svcCtx := &svc.ServiceContext{
				BalanceModel: mockBalanceModel,
			}

			// 创建逻辑实例
			ctx := tt.setupContext()
			logic := NewGetBalancesLogic(ctx, svcCtx)

			// 执行测试
			result, err := logic.GetBalances()

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, len(tt.expectedResult.Balances), len(result.Balances))
				
				// 验证每个余额记录
				for i, expectedBalance := range tt.expectedResult.Balances {
					assert.Equal(t, expectedBalance.Currency, result.Balances[i].Currency)
					assert.Equal(t, expectedBalance.Available, result.Balances[i].Available)
					assert.Equal(t, expectedBalance.Frozen, result.Balances[i].Frozen)
					assert.Equal(t, expectedBalance.UpdatedAt, result.Balances[i].UpdatedAt)
				}
			}

			// 验证模拟调用
			mockBalanceModel.AssertExpectations(t)
		})
	}
}

func TestGetBalancesLogic_getUserIDFromContext(t *testing.T) {
	tests := []struct {
		name           string
		contextValue   interface{}
		expectedUserID uint64
		expectedError  bool
	}{
		{
			name:           "float64类型的用户ID",
			contextValue:   float64(123),
			expectedUserID: 123,
			expectedError:  false,
		},
		{
			name:           "uint64类型的用户ID",
			contextValue:   uint64(456),
			expectedUserID: 456,
			expectedError:  false,
		},
		{
			name:           "int64类型的用户ID",
			contextValue:   int64(789),
			expectedUserID: 789,
			expectedError:  false,
		},
		{
			name:           "int类型的用户ID",
			contextValue:   int(101112),
			expectedUserID: 101112,
			expectedError:  false,
		},
		{
			name:           "无效类型的用户ID",
			contextValue:   "invalid",
			expectedUserID: 0,
			expectedError:  true,
		},
		{
			name:           "上下文中无用户ID",
			contextValue:   nil,
			expectedUserID: 0,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建上下文
			var ctx context.Context
			if tt.contextValue != nil {
				ctx = context.WithValue(context.Background(), "userId", tt.contextValue)
			} else {
				ctx = context.Background()
			}

			// 创建逻辑实例
			logic := NewGetBalancesLogic(ctx, &svc.ServiceContext{})

			// 执行测试
			userID, err := logic.getUserIDFromContext()

			// 验证结果
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, uint64(0), userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserID, userID)
			}
		})
	}
}

func TestGetBalancesLogic_validateBalanceAmounts(t *testing.T) {
	logic := NewGetBalancesLogic(context.Background(), &svc.ServiceContext{})

	tests := []struct {
		name          string
		available     string
		frozen        string
		expectedError bool
	}{
		{
			name:          "有效的余额数值",
			available:     "100.50000000",
			frozen:        "25.25000000",
			expectedError: false,
		},
		{
			name:          "零余额",
			available:     "0.00000000",
			frozen:        "0.00000000",
			expectedError: false,
		},
		{
			name:          "无效的可用余额格式",
			available:     "invalid",
			frozen:        "0.00000000",
			expectedError: true,
		},
		{
			name:          "无效的冻结余额格式",
			available:     "100.00000000",
			frozen:        "invalid",
			expectedError: true,
		},
		{
			name:          "负数可用余额",
			available:     "-100.00000000",
			frozen:        "0.00000000",
			expectedError: true,
		},
		{
			name:          "负数冻结余额",
			available:     "100.00000000",
			frozen:        "-25.00000000",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logic.validateBalanceAmounts(tt.available, tt.frozen)
			
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}