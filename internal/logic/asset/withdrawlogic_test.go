package asset

import (
	"context"
	"strings"
	"testing"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/logx"
)

// 使用 depositlogic_test.go 中的 MockBalanceModel 和 MockAssetTransactionModel



func TestWithdrawLogic_Withdraw(t *testing.T) {
	// 禁用日志输出以保持测试输出清洁
	logx.DisableStat()

	tests := []struct {
		name           string
		userID         uint64
		request        *types.WithdrawRequest
		existingBalance *model.Balance
		balanceError   error
		updateError    error
		expectedError  error
		setupContext   func() context.Context
		validateResult func(t *testing.T, resp *types.WithdrawResponse)
	}{
		{
			name:   "成功提现BTC",
			userID: 1,
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			existingBalance: &model.Balance{
				ID:        1,
				UserID:    1,
				Currency:  "BTC",
				Available: "2.00000000", // 足够的余额
				Frozen:    "0.00000000",
				UpdatedAt: time.Now(),
			},
			balanceError:  nil,
			updateError:   nil,
			expectedError: nil,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(1))
			},
			validateResult: func(t *testing.T, resp *types.WithdrawResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, "BTC", resp.Currency)
				assert.Equal(t, "1.00000000", resp.Amount)
				assert.Equal(t, "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", resp.Address)
				assert.Equal(t, "0.0005", resp.Fee) // BTC手续费
				assert.Equal(t, int64(1), resp.Status) // 待审核状态
				assert.NotEmpty(t, resp.TransactionID)
				assert.Contains(t, resp.TransactionID, "WTH_")
				assert.NotEmpty(t, resp.CreatedAt)
			},
		},
		{
			name:   "成功提现USDT",
			userID: 2,
			request: &types.WithdrawRequest{
				Currency: "USDT",
				Amount:   "100.00000000",
				Address:  "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
			},
			existingBalance: &model.Balance{
				ID:        2,
				UserID:    2,
				Currency:  "USDT",
				Available: "500.00000000", // 足够的余额
				Frozen:    "0.00000000",
				UpdatedAt: time.Now(),
			},
			balanceError:  nil,
			updateError:   nil,
			expectedError: nil,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(2))
			},
			validateResult: func(t *testing.T, resp *types.WithdrawResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, "USDT", resp.Currency)
				assert.Equal(t, "100.00000000", resp.Amount)
				assert.Equal(t, "1", resp.Fee) // USDT手续费
				assert.Equal(t, int64(1), resp.Status)
			},
		},
		{
			name:   "余额不足",
			userID: 3,
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			existingBalance: &model.Balance{
				ID:        3,
				UserID:    3,
				Currency:  "BTC",
				Available: "0.50000000", // 余额不足（需要1.0005包含手续费）
				Frozen:    "0.00000000",
				UpdatedAt: time.Now(),
			},
			balanceError:  nil,
			updateError:   nil,
			expectedError: model.ErrInsufficientBalance,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(3))
			},
			validateResult: nil,
		},
		{
			name:   "余额记录不存在",
			userID: 4,
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			existingBalance: nil,
			balanceError:   model.ErrNotFound,
			updateError:    nil,
			expectedError:  model.ErrBalanceNotFound,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(4))
			},
			validateResult: nil,
		},
		{
			name:   "无效的提现金额",
			userID: 5,
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "invalid_amount",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			existingBalance: nil,
			balanceError:   nil,
			updateError:    nil,
			expectedError:  model.ErrInvalidAmount,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(5))
			},
			validateResult: nil,
		},
		{
			name:   "负数提现金额",
			userID: 6,
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "-1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			existingBalance: nil,
			balanceError:   nil,
			updateError:    nil,
			expectedError:  model.ErrInvalidAmount,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(6))
			},
			validateResult: nil,
		},
		{
			name:   "零提现金额",
			userID: 7,
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "0.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			existingBalance: nil,
			balanceError:   nil,
			updateError:    nil,
			expectedError:  model.ErrInvalidAmount,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(7))
			},
			validateResult: nil,
		},
		{
			name:   "不支持的币种",
			userID: 8,
			request: &types.WithdrawRequest{
				Currency: "INVALID",
				Amount:   "1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			existingBalance: nil,
			balanceError:   nil,
			updateError:    nil,
			expectedError:  model.ErrCurrencyNotFound,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(8))
			},
			validateResult: nil,
		},
		{
			name:   "无效的BTC地址",
			userID: 9,
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
				Address:  "invalid_btc_address",
			},
			existingBalance: nil,
			balanceError:   nil,
			updateError:    nil,
			expectedError:  assert.AnError, // 地址格式错误
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(9))
			},
			validateResult: nil,
		},
		{
			name:   "无效的ETH地址",
			userID: 10,
			request: &types.WithdrawRequest{
				Currency: "ETH",
				Amount:   "1.00000000",
				Address:  "invalid_eth_address",
			},
			existingBalance: nil,
			balanceError:   nil,
			updateError:    nil,
			expectedError:  assert.AnError, // 地址格式错误
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(10))
			},
			validateResult: nil,
		},
		{
			name:   "低于最小提现金额",
			userID: 11,
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "0.0001", // 低于最小值0.001
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			existingBalance: nil,
			balanceError:   nil,
			updateError:    nil,
			expectedError:  assert.AnError, // 最小金额错误
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(11))
			},
			validateResult: nil,
		},
		{
			name:   "JWT上下文中无用户ID",
			userID: 0,
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			existingBalance: nil,
			balanceError:   nil,
			updateError:    nil,
			expectedError:  model.ErrUnauthorized,
			setupContext: func() context.Context {
				return context.Background() // 没有userId
			},
			validateResult: nil,
		},
		{
			name:   "数据库更新错误",
			userID: 12,
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			existingBalance: &model.Balance{
				ID:        12,
				UserID:    12,
				Currency:  "BTC",
				Available: "2.00000000",
				Frozen:    "0.00000000",
				UpdatedAt: time.Now(),
			},
			balanceError:  nil,
			updateError:   assert.AnError,
			expectedError: assert.AnError,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userId", float64(12))
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
			shouldSetupDatabaseMocks := tt.userID > 0 && 
				tt.expectedError != model.ErrInvalidAmount && 
				tt.expectedError != model.ErrCurrencyNotFound && 
				tt.expectedError != model.ErrInvalidParams &&
				tt.expectedError != model.ErrUnauthorized &&
				!strings.Contains(tt.name, "无效的") &&
				!strings.Contains(tt.name, "低于最小")
			
			if shouldSetupDatabaseMocks {
				// 只有在通过验证的情况下才设置数据库操作的mock
				
				// 模拟Trans方法调用
				mockBalanceModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(nil)

				// 模拟FindByUserIDAndCurrency调用
				if tt.existingBalance != nil {
					mockBalanceModel.On("FindByUserIDAndCurrency", mock.Anything, tt.userID, tt.request.Currency).Return(tt.existingBalance, tt.balanceError)
				} else {
					mockBalanceModel.On("FindByUserIDAndCurrency", mock.Anything, tt.userID, tt.request.Currency).Return((*model.Balance)(nil), tt.balanceError)
				}

				// 如果有现有余额且没有更新错误，设置UpdateBalance期望
				if tt.existingBalance != nil && tt.updateError == nil && tt.expectedError == nil {
					mockBalanceModel.On("UpdateBalance", mock.Anything, tt.userID, tt.request.Currency, mock.AnythingOfType("string"), tt.existingBalance.Frozen).Return(tt.updateError)
				} else if tt.existingBalance != nil && tt.updateError != nil {
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
			logic := NewWithdrawLogic(ctx, svcCtx)

			// 执行测试
			result, err := logic.Withdraw(tt.request)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError == assert.AnError {
					// 对于assert.AnError，检查错误消息或类型
					assert.NotNil(t, err)
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

func TestWithdrawLogic_validateWithdrawRequest(t *testing.T) {
	logic := NewWithdrawLogic(context.Background(), &svc.ServiceContext{})

	tests := []struct {
		name          string
		request       *types.WithdrawRequest
		expectedError bool
	}{
		{
			name: "有效的BTC提现请求",
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: false,
		},
		{
			name: "有效的ETH提现请求",
			request: &types.WithdrawRequest{
				Currency: "ETH",
				Amount:   "1.00000000",
				Address:  "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
			},
			expectedError: false,
		},
		{
			name: "小写币种代码自动转换",
			request: &types.WithdrawRequest{
				Currency: "btc",
				Amount:   "1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: false,
		},
		{
			name: "空币种代码",
			request: &types.WithdrawRequest{
				Currency: "",
				Amount:   "1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: true,
		},
		{
			name: "不支持的币种",
			request: &types.WithdrawRequest{
				Currency: "INVALID",
				Amount:   "1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: true,
		},
		{
			name: "空提现金额",
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: true,
		},
		{
			name: "无效的提现金额格式",
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "invalid",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: true,
		},
		{
			name: "负数提现金额",
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "-1.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: true,
		},
		{
			name: "零提现金额",
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "0.00000000",
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: true,
		},
		{
			name: "精度超过8位小数",
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "1.123456789", // 9位小数
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: true,
		},
		{
			name: "低于最小提现金额",
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "0.0001", // 低于BTC最小值0.001
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: true,
		},
		{
			name: "超过最大提现金额",
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "100.00000000", // 超过BTC最大值10.0
				Address:  "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			},
			expectedError: true,
		},
		{
			name: "空提现地址",
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
				Address:  "",
			},
			expectedError: true,
		},
		{
			name: "无效的BTC地址格式",
			request: &types.WithdrawRequest{
				Currency: "BTC",
				Amount:   "1.00000000",
				Address:  "invalid_btc_address",
			},
			expectedError: true,
		},
		{
			name: "无效的ETH地址格式",
			request: &types.WithdrawRequest{
				Currency: "ETH",
				Amount:   "1.00000000",
				Address:  "invalid_eth_address",
			},
			expectedError: true,
		},
		{
			name: "ETH地址长度不正确",
			request: &types.WithdrawRequest{
				Currency: "ETH",
				Amount:   "1.00000000",
				Address:  "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8", // 少了2个字符
			},
			expectedError: true,
		},
		{
			name: "ETH地址不以0x开头",
			request: &types.WithdrawRequest{
				Currency: "ETH",
				Amount:   "1.00000000",
				Address:  "742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logic.validateWithdrawRequest(tt.request)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// 验证币种代码被转换为大写
				assert.Equal(t, strings.ToUpper(tt.request.Currency), tt.request.Currency)
			}
		})
	}
}

func TestWithdrawLogic_calculateWithdrawFee(t *testing.T) {
	logic := NewWithdrawLogic(context.Background(), &svc.ServiceContext{})

	tests := []struct {
		name         string
		currency     string
		amount       string
		expectedFee  string
	}{
		{
			name:        "BTC手续费",
			currency:    "BTC",
			amount:      "1.00000000",
			expectedFee: "0.0005",
		},
		{
			name:        "ETH手续费",
			currency:    "ETH",
			amount:      "10.00000000",
			expectedFee: "0.005",
		},
		{
			name:        "USDT手续费",
			currency:    "USDT",
			amount:      "100.00000000",
			expectedFee: "1",
		},
		{
			name:        "USDC手续费",
			currency:    "USDC",
			amount:      "100.00000000",
			expectedFee: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, _ := decimal.NewFromString(tt.amount)
			fee := logic.calculateWithdrawFee(tt.currency, amount)
			assert.Equal(t, tt.expectedFee, fee.String())
		})
	}
}

func TestWithdrawLogic_generateTransactionID(t *testing.T) {
	logic := NewWithdrawLogic(context.Background(), &svc.ServiceContext{})

	// 生成多个交易ID并验证格式和唯一性
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := logic.generateTransactionID()
		
		// 验证格式
		assert.True(t, strings.HasPrefix(id, "WTH_"))
		assert.Equal(t, 36, len(id)) // WTH_ (4) + UUID without dashes (32)
		
		// 验证唯一性
		assert.False(t, ids[id], "Transaction ID should be unique: %s", id)
		ids[id] = true
	}
}