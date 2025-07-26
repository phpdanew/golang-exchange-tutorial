package asset

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// MockBalanceModel 模拟BalanceModel接口
type MockBalanceModel struct {
	mock.Mock
}

func (m *MockBalanceModel) Insert(ctx context.Context, data *model.Balance) (sql.Result, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(sql.Result), args.Error(1)
}

func (m *MockBalanceModel) FindOne(ctx context.Context, id uint64) (*model.Balance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Balance), args.Error(1)
}

func (m *MockBalanceModel) FindByUserID(ctx context.Context, userID uint64) ([]*model.Balance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Balance), args.Error(1)
}

func (m *MockBalanceModel) FindByUserIDAndCurrency(ctx context.Context, userID uint64, currency string) (*model.Balance, error) {
	args := m.Called(ctx, userID, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Balance), args.Error(1)
}

func (m *MockBalanceModel) UpdateBalance(ctx context.Context, userID uint64, currency string, available, frozen string) error {
	args := m.Called(ctx, userID, currency, available, frozen)
	return args.Error(0)
}

func (m *MockBalanceModel) FreezeBalance(ctx context.Context, userID uint64, currency string, amount string) error {
	args := m.Called(ctx, userID, currency, amount)
	return args.Error(0)
}

func (m *MockBalanceModel) UnfreezeBalance(ctx context.Context, userID uint64, currency string, amount string) error {
	args := m.Called(ctx, userID, currency, amount)
	return args.Error(0)
}

func (m *MockBalanceModel) Update(ctx context.Context, data *model.Balance) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockBalanceModel) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBalanceModel) Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	args := m.Called(ctx, fn)
	// 执行传入的函数并返回其结果
	if fn != nil {
		return fn(ctx, nil)
	}
	return args.Error(0)
}

// MockAssetTransactionModel 模拟AssetTransactionModel接口
type MockAssetTransactionModel struct {
	mock.Mock
}

func (m *MockAssetTransactionModel) Insert(ctx context.Context, data *model.AssetTransaction) (sql.Result, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(sql.Result), args.Error(1)
}

func (m *MockAssetTransactionModel) FindOne(ctx context.Context, id uint64) (*model.AssetTransaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AssetTransaction), args.Error(1)
}

func (m *MockAssetTransactionModel) FindByUserID(ctx context.Context, userID uint64, limit, offset int64) ([]*model.AssetTransaction, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AssetTransaction), args.Error(1)
}

func (m *MockAssetTransactionModel) FindByTransactionID(ctx context.Context, transactionID string) (*model.AssetTransaction, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AssetTransaction), args.Error(1)
}

func (m *MockAssetTransactionModel) FindByUserIDAndType(ctx context.Context, userID uint64, transactionType int64, limit, offset int64) ([]*model.AssetTransaction, error) {
	args := m.Called(ctx, userID, transactionType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AssetTransaction), args.Error(1)
}

func (m *MockAssetTransactionModel) CountByUserID(ctx context.Context, userID uint64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAssetTransactionModel) CountByUserIDAndType(ctx context.Context, userID uint64, transactionType int64) (int64, error) {
	args := m.Called(ctx, userID, transactionType)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAssetTransactionModel) Update(ctx context.Context, data *model.AssetTransaction) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockAssetTransactionModel) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestDepositLogic_Deposit(t *testing.T) {
	// 禁用日志输出以保持测试输出清洁
	logx.DisableStat()

	tests := []struct {
		name           string
		userID         uint64
		request        *types.DepositRequest
		existingBalance *model.Balance
		balanceError   error
		insertError    error
		updateError    error
		expectedError  error
		setupContext   func() context.Context
		validateResult func(t *testing.T, resp *types.DepositResponse)
	}{
		{
			name:   "成功充值到新币种",
			userID: 1,
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "1.50000000",
			},
			existingBalance: nil,
			balanceError:   model.ErrNotFound,
			insertError:    nil,
			updateError:    nil,
			expectedError:  nil,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(1))
			},
			validateResult: func(t *testing.T, resp *types.DepositResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, "BTC", resp.Currency)
				assert.Equal(t, "1.50000000", resp.Amount)
				assert.Equal(t, int64(2), resp.Status) // 成功状态
				assert.NotEmpty(t, resp.TransactionID)
				assert.Contains(t, resp.TransactionID, "DEP_")
				assert.NotEmpty(t, resp.CreatedAt)
			},
		},
		{
			name:   "成功充值到现有币种",
			userID: 2,
			request: &types.DepositRequest{
				Currency: "USDT",
				Amount:   "1000.00000000",
			},
			existingBalance: &model.Balance{
				ID:        1,
				UserID:    2,
				Currency:  "USDT",
				Available: "500.00000000",
				Frozen:    "100.00000000",
				UpdatedAt: time.Now(),
			},
			balanceError:  nil,
			insertError:   nil,
			updateError:   nil,
			expectedError: nil,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(2))
			},
			validateResult: func(t *testing.T, resp *types.DepositResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, "USDT", resp.Currency)
				assert.Equal(t, "1000.00000000", resp.Amount)
				assert.Equal(t, int64(2), resp.Status)
				assert.NotEmpty(t, resp.TransactionID)
			},
		},
		{
			name:   "无效的充值金额",
			userID: 3,
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "invalid_amount",
			},
			existingBalance: nil,
			balanceError:   nil,
			insertError:    nil,
			updateError:    nil,
			expectedError:  model.ErrInvalidAmount,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(3))
			},
			validateResult: nil,
		},
		{
			name:   "负数充值金额",
			userID: 4,
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "-1.00000000",
			},
			existingBalance: nil,
			balanceError:   nil,
			insertError:    nil,
			updateError:    nil,
			expectedError:  model.ErrInvalidAmount,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(4))
			},
			validateResult: nil,
		},
		{
			name:   "零充值金额",
			userID: 5,
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "0.00000000",
			},
			existingBalance: nil,
			balanceError:   nil,
			insertError:    nil,
			updateError:    nil,
			expectedError:  model.ErrInvalidAmount,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(5))
			},
			validateResult: nil,
		},
		{
			name:   "不支持的币种",
			userID: 6,
			request: &types.DepositRequest{
				Currency: "INVALID",
				Amount:   "1.00000000",
			},
			existingBalance: nil,
			balanceError:   nil,
			insertError:    nil,
			updateError:    nil,
			expectedError:  model.ErrCurrencyNotFound,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(6))
			},
			validateResult: nil,
		},
		{
			name:   "JWT上下文中无用户ID",
			userID: 0,
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
			},
			existingBalance: nil,
			balanceError:   nil,
			insertError:    nil,
			updateError:    nil,
			expectedError:  model.ErrUnauthorized,
			setupContext: func() context.Context {
				return context.Background() // 没有userId
			},
			validateResult: nil,
		},
		{
			name:   "数据库插入错误",
			userID: 7,
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
			},
			existingBalance: nil,
			balanceError:   model.ErrNotFound,
			insertError:    assert.AnError,
			updateError:    nil,
			expectedError:  assert.AnError,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(7))
			},
			validateResult: nil,
		},
		{
			name:   "数据库更新错误",
			userID: 8,
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
			},
			existingBalance: &model.Balance{
				ID:        1,
				UserID:    8,
				Currency:  "BTC",
				Available: "1.00000000",
				Frozen:    "0.00000000",
				UpdatedAt: time.Now(),
			},
			balanceError:  nil,
			insertError:   nil,
			updateError:   assert.AnError,
			expectedError: assert.AnError,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(8))
			},
			validateResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟的BalanceModel和AssetTransactionModel
			mockBalanceModel := new(MockBalanceModel)
			mockAssetTransactionModel := new(MockAssetTransactionModel)

			// 设置模拟期望
			if tt.userID > 0 && tt.expectedError != model.ErrInvalidAmount && tt.expectedError != model.ErrCurrencyNotFound && tt.expectedError != model.ErrInvalidParams {
				// 只有在通过验证的情况下才设置数据库操作的mock
				
				// 模拟Trans方法调用 - 需要实际执行传入的函数
				var transError error
				if tt.insertError != nil {
					transError = tt.insertError
				} else if tt.updateError != nil {
					transError = tt.updateError
				}
				
				mockBalanceModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(transError).Run(func(args mock.Arguments) {
					// 获取传入的函数并执行
					fn := args.Get(1).(func(context.Context, sqlx.Session) error)
					fn(args.Get(0).(context.Context), nil) // 传入nil作为session
				})

				// 模拟FindByUserIDAndCurrency调用
				if tt.existingBalance != nil {
					mockBalanceModel.On("FindByUserIDAndCurrency", mock.Anything, tt.userID, tt.request.Currency).Return(tt.existingBalance, tt.balanceError)
				} else {
					mockBalanceModel.On("FindByUserIDAndCurrency", mock.Anything, tt.userID, tt.request.Currency).Return((*model.Balance)(nil), tt.balanceError)
				}

				// 根据是否存在余额记录来设置不同的期望
				if tt.balanceError != nil && tt.balanceError == model.ErrNotFound {
					// 新建余额记录
					mockBalanceModel.On("Insert", mock.Anything, mock.AnythingOfType("*model.Balance")).Return(nil, tt.insertError)
				} else if tt.existingBalance != nil {
					// 更新现有余额记录
					mockBalanceModel.On("UpdateBalance", mock.Anything, tt.userID, tt.request.Currency, mock.AnythingOfType("string"), tt.existingBalance.Frozen).Return(tt.updateError)
				}

				// 设置AssetTransactionModel的mock期望
				if tt.expectedError == nil {
					// 成功情况下，应该创建交易记录
					mockAssetTransactionModel.On("Insert", mock.Anything, mock.AnythingOfType("*model.AssetTransaction")).Return(nil, nil)
				}
			}

			// 创建服务上下文
			svcCtx := &svc.ServiceContext{
				BalanceModel:          mockBalanceModel,
				AssetTransactionModel: mockAssetTransactionModel,
			}

			// 创建逻辑实例
			ctx := tt.setupContext()
			logic := NewDepositLogic(ctx, svcCtx)

			// 执行测试
			result, err := logic.Deposit(tt.request)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError == assert.AnError {
					// 对于assert.AnError，检查错误消息
					assert.Contains(t, err.Error(), "assert.AnError")
				} else {
					assert.Equal(t, tt.expectedError, err)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			// 验证模拟调用
			mockBalanceModel.AssertExpectations(t)
			mockAssetTransactionModel.AssertExpectations(t)
		})
	}
}

func TestDepositLogic_validateDepositRequest(t *testing.T) {
	logic := NewDepositLogic(context.Background(), &svc.ServiceContext{})

	tests := []struct {
		name          string
		request       *types.DepositRequest
		expectedError error
	}{
		{
			name: "有效的充值请求",
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "1.50000000",
			},
			expectedError: nil,
		},
		{
			name: "小写币种代码自动转换",
			request: &types.DepositRequest{
				Currency: "btc",
				Amount:   "1.00000000",
			},
			expectedError: nil,
		},
		{
			name: "空币种代码",
			request: &types.DepositRequest{
				Currency: "",
				Amount:   "1.00000000",
			},
			expectedError: model.ErrInvalidParams,
		},
		{
			name: "不支持的币种",
			request: &types.DepositRequest{
				Currency: "INVALID",
				Amount:   "1.00000000",
			},
			expectedError: model.ErrCurrencyNotFound,
		},
		{
			name: "空充值金额",
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "",
			},
			expectedError: model.ErrInvalidAmount,
		},
		{
			name: "无效的充值金额格式",
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "invalid",
			},
			expectedError: model.ErrInvalidAmount,
		},
		{
			name: "负数充值金额",
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "-1.00000000",
			},
			expectedError: model.ErrInvalidAmount,
		},
		{
			name: "零充值金额",
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "0.00000000",
			},
			expectedError: model.ErrInvalidAmount,
		},
		{
			name: "精度超过8位小数",
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "1.123456789", // 9位小数
			},
			expectedError: assert.AnError, // 精度错误
		},
		{
			name: "低于最小充值金额",
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "0.000000001", // 小于最小值
			},
			expectedError: assert.AnError, // 最小金额错误
		},
		{
			name: "超过最大充值金额",
			request: &types.DepositRequest{
				Currency: "BTC",
				Amount:   "10000000.00000000", // 超过最大值
			},
			expectedError: assert.AnError, // 最大金额错误
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logic.validateDepositRequest(tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				// 对于特定的错误类型，进行精确匹配
				if tt.expectedError != assert.AnError {
					assert.Equal(t, tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err)
				// 验证币种代码被转换为大写
				assert.Equal(t, strings.ToUpper(tt.request.Currency), tt.request.Currency)
			}
		})
	}
}

func TestDepositLogic_generateTransactionID(t *testing.T) {
	logic := NewDepositLogic(context.Background(), &svc.ServiceContext{})

	// 生成多个交易ID并验证格式和唯一性
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := logic.generateTransactionID()
		
		// 验证格式
		assert.True(t, strings.HasPrefix(id, "DEP_"))
		assert.Equal(t, 36, len(id)) // DEP_ (4) + UUID without dashes (32)
		
		// 验证唯一性
		assert.False(t, ids[id], "Transaction ID should be unique: %s", id)
		ids[id] = true
	}
}