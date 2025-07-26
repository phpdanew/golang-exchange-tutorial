package trading

import (
	"context"
	"strconv"
	"testing"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelOrderLogic_CancelOrder(t *testing.T) {
	tests := []struct {
		name            string
		request         *types.CancelOrderRequest
		userID          uint64
		order           *model.Order
		orderErr        error
		tradingPair     *model.TradingPair
		tradingPairErr  error
		unfreezeErr     error
		updateStatusErr error
		expectedError   string
		expectedUnfreeze struct {
			currency string
			amount   string
		}
	}{
		{
			name: "成功取消限价买单",
			request: &types.CancelOrderRequest{
				OrderID: 123,
			},
			userID: 1,
			order: &model.Order{
				ID:           123,
				UserID:       1,
				Symbol:       "BTC/USDT",
				Type:         1, // 限价单
				Side:         1, // 买入
				Amount:       "1.00000000",
				Price:        "50000.00000000",
				FilledAmount: "0.30000000", // 部分成交
				Status:       2,            // 部分成交状态
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			orderErr: nil,
			tradingPair: &model.TradingPair{
				ID:            1,
				Symbol:        "BTC/USDT",
				BaseCurrency:  "BTC",
				QuoteCurrency: "USDT",
				Status:        1,
			},
			tradingPairErr:  nil,
			unfreezeErr:     nil,
			updateStatusErr: nil,
			expectedError:   "",
			expectedUnfreeze: struct {
				currency string
				amount   string
			}{
				currency: "USDT",
				amount:   "35000", // (1.0 - 0.3) * 50000 = 0.7 * 50000
			},
		},
		{
			name: "成功取消限价卖单",
			request: &types.CancelOrderRequest{
				OrderID: 124,
			},
			userID: 2,
			order: &model.Order{
				ID:           124,
				UserID:       2,
				Symbol:       "BTC/USDT",
				Type:         1, // 限价单
				Side:         2, // 卖出
				Amount:       "2.00000000",
				Price:        "51000.00000000",
				FilledAmount: "0.50000000", // 部分成交
				Status:       2,            // 部分成交状态
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			orderErr: nil,
			tradingPair: &model.TradingPair{
				ID:            1,
				Symbol:        "BTC/USDT",
				BaseCurrency:  "BTC",
				QuoteCurrency: "USDT",
				Status:        1,
			},
			tradingPairErr:  nil,
			unfreezeErr:     nil,
			updateStatusErr: nil,
			expectedError:   "",
			expectedUnfreeze: struct {
				currency string
				amount   string
			}{
				currency: "BTC",
				amount:   "1.5", // 2.0 - 0.5 = 1.5 (剩余未成交的BTC)
			},
		},
		{
			name: "成功取消完全未成交的订单",
			request: &types.CancelOrderRequest{
				OrderID: 125,
			},
			userID: 3,
			order: &model.Order{
				ID:           125,
				UserID:       3,
				Symbol:       "BTC/USDT",
				Type:         1, // 限价单
				Side:         1, // 买入
				Amount:       "0.50000000",
				Price:        "49000.00000000",
				FilledAmount: "0.00000000", // 完全未成交
				Status:       1,            // 待成交状态
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			orderErr: nil,
			tradingPair: &model.TradingPair{
				ID:            1,
				Symbol:        "BTC/USDT",
				BaseCurrency:  "BTC",
				QuoteCurrency: "USDT",
				Status:        1,
			},
			tradingPairErr:  nil,
			unfreezeErr:     nil,
			updateStatusErr: nil,
			expectedError:   "",
			expectedUnfreeze: struct {
				currency string
				amount   string
			}{
				currency: "USDT",
				amount:   "24500", // 0.5 * 49000
			},
		},
		{
			name: "订单不存在",
			request: &types.CancelOrderRequest{
				OrderID: 999,
			},
			userID:          4,
			order:           nil,
			orderErr:        model.ErrNotFound,
			tradingPair:     nil,
			tradingPairErr:  nil,
			unfreezeErr:     nil,
			updateStatusErr: nil,
			expectedError:   "order not found",
		},
		{
			name: "无权限取消他人订单",
			request: &types.CancelOrderRequest{
				OrderID: 126,
			},
			userID: 5,
			order: &model.Order{
				ID:           126,
				UserID:       999, // 不同的用户ID
				Symbol:       "BTC/USDT",
				Type:         1,
				Side:         1,
				Amount:       "1.00000000",
				Price:        "50000.00000000",
				FilledAmount: "0.00000000",
				Status:       1,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			orderErr:        nil,
			tradingPair:     nil,
			tradingPairErr:  nil,
			unfreezeErr:     nil,
			updateStatusErr: nil,
			expectedError:   "forbidden",
		},
		{
			name: "订单已取消",
			request: &types.CancelOrderRequest{
				OrderID: 127,
			},
			userID: 6,
			order: &model.Order{
				ID:           127,
				UserID:       6,
				Symbol:       "BTC/USDT",
				Type:         1,
				Side:         1,
				Amount:       "1.00000000",
				Price:        "50000.00000000",
				FilledAmount: "0.00000000",
				Status:       4, // 已取消状态
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			orderErr:        nil,
			tradingPair:     nil,
			tradingPairErr:  nil,
			unfreezeErr:     nil,
			updateStatusErr: nil,
			expectedError:   "order already canceled",
		},
		{
			name: "订单已完全成交",
			request: &types.CancelOrderRequest{
				OrderID: 128,
			},
			userID: 7,
			order: &model.Order{
				ID:           128,
				UserID:       7,
				Symbol:       "BTC/USDT",
				Type:         1,
				Side:         1,
				Amount:       "1.00000000",
				Price:        "50000.00000000",
				FilledAmount: "1.00000000", // 完全成交
				Status:       3,            // 完全成交状态
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			orderErr:        nil,
			tradingPair:     nil,
			tradingPairErr:  nil,
			unfreezeErr:     nil,
			updateStatusErr: nil,
			expectedError:   "order already filled",
		},
		{
			name: "完全成交的订单无需解冻",
			request: &types.CancelOrderRequest{
				OrderID: 129,
			},
			userID: 8,
			order: &model.Order{
				ID:           129,
				UserID:       8,
				Symbol:       "BTC/USDT",
				Type:         1, // 限价单
				Side:         1, // 买入
				Amount:       "1.00000000",
				Price:        "50000.00000000",
				FilledAmount: "1.00000000", // 完全成交，但状态还是部分成交（边界情况）
				Status:       2,            // 部分成交状态
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			orderErr: nil,
			tradingPair: &model.TradingPair{
				ID:            1,
				Symbol:        "BTC/USDT",
				BaseCurrency:  "BTC",
				QuoteCurrency: "USDT",
				Status:        1,
			},
			tradingPairErr:  nil,
			unfreezeErr:     nil,
			updateStatusErr: nil,
			expectedError:   "",
			expectedUnfreeze: struct {
				currency string
				amount   string
			}{
				currency: "",
				amount:   "0", // 无需解冻
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建mock对象
			mockOrderModel := &mockOrderModel{}
			mockTradingPairModel := &mockTradingPairModel{}
			mockBalanceModel := &mockBalanceModel{}

			// 设置mock期望
			mockOrderModel.On("FindOne", mock.Anything, tt.request.OrderID).Return(tt.order, tt.orderErr)

			if tt.orderErr == nil && tt.order != nil && tt.order.UserID == tt.userID && tt.order.Status != 4 && tt.order.Status != 3 {
				mockTradingPairModel.On("FindBySymbol", mock.Anything, tt.order.Symbol).Return(tt.tradingPair, tt.tradingPairErr)

				if tt.tradingPairErr == nil {
					mockOrderModel.On("UpdateStatus", mock.Anything, tt.order.ID, int64(4)).Return(tt.updateStatusErr)
					mockOrderModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(nil)

					// 只有当需要解冻时才设置解冻期望
					if tt.expectedUnfreeze.amount != "0" && tt.expectedUnfreeze.currency != "" {
						mockBalanceModel.On("UnfreezeBalance", mock.Anything, tt.userID, tt.expectedUnfreeze.currency, tt.expectedUnfreeze.amount).Return(tt.unfreezeErr)
					}
				}
			}

			// 创建服务上下文
			svcCtx := &svc.ServiceContext{
				OrderModel:       mockOrderModel,
				TradingPairModel: mockTradingPairModel,
				BalanceModel:     mockBalanceModel,
			}

			// 创建带用户ID的上下文
			ctx := context.WithValue(context.Background(), "userId", strconv.FormatUint(tt.userID, 10))

			// 创建逻辑实例
			logic := NewCancelOrderLogic(ctx, svcCtx)

			// 执行测试
			resp, err := logic.CancelOrder(tt.request)

			// 验证结果
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, 0, resp.Code)
				assert.Equal(t, "Order canceled successfully", resp.Message)
			}

			// 验证mock调用
			mockOrderModel.AssertExpectations(t)
			mockTradingPairModel.AssertExpectations(t)
			mockBalanceModel.AssertExpectations(t)
		})
	}
}

func TestCancelOrderLogic_calculateUnfreezeAmount(t *testing.T) {
	logic := &CancelOrderLogic{}

	tradingPair := &model.TradingPair{
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
	}

	tests := []struct {
		name             string
		order            *model.Order
		expectedCurrency string
		expectedAmount   string
		expectedError    string
	}{
		{
			name: "限价买单部分成交解冻计算",
			order: &model.Order{
				Type:         1, // 限价单
				Side:         1, // 买入
				Amount:       "2.00000000",
				Price:        "45000.00000000",
				FilledAmount: "0.75000000", // 部分成交
			},
			expectedCurrency: "USDT",
			expectedAmount:   "56250", // (2.0 - 0.75) * 45000 = 1.25 * 45000
			expectedError:    "",
		},
		{
			name: "限价卖单部分成交解冻计算",
			order: &model.Order{
				Type:         1, // 限价单
				Side:         2, // 卖出
				Amount:       "1.50000000",
				Price:        "46000.00000000",
				FilledAmount: "0.20000000", // 部分成交
			},
			expectedCurrency: "BTC",
			expectedAmount:   "1.3", // 1.5 - 0.2 = 1.3 (剩余未成交的BTC)
			expectedError:    "",
		},
		{
			name: "市价买单部分成交解冻计算",
			order: &model.Order{
				Type:         2, // 市价单
				Side:         1, // 买入
				Amount:       "10000.00000000", // 市价买单Amount表示要花费的USDT
				FilledAmount: "3000.00000000",  // 已花费的USDT
			},
			expectedCurrency: "USDT",
			expectedAmount:   "7000", // 10000 - 3000 = 7000 (剩余USDT)
			expectedError:    "",
		},
		{
			name: "市价卖单部分成交解冻计算",
			order: &model.Order{
				Type:         2, // 市价单
				Side:         2, // 卖出
				Amount:       "0.80000000",
				FilledAmount: "0.30000000", // 部分成交
			},
			expectedCurrency: "BTC",
			expectedAmount:   "0.5", // 0.8 - 0.3 = 0.5 (剩余BTC)
			expectedError:    "",
		},
		{
			name: "完全成交无需解冻",
			order: &model.Order{
				Type:         1, // 限价单
				Side:         1, // 买入
				Amount:       "1.00000000",
				Price:        "50000.00000000",
				FilledAmount: "1.00000000", // 完全成交
			},
			expectedCurrency: "",
			expectedAmount:   "0", // 无需解冻
			expectedError:    "",
		},
		{
			name: "超额成交处理（边界情况）",
			order: &model.Order{
				Type:         1, // 限价单
				Side:         1, // 买入
				Amount:       "1.00000000",
				Price:        "50000.00000000",
				FilledAmount: "1.10000000", // 超额成交（理论上不应该发生）
			},
			expectedCurrency: "",
			expectedAmount:   "0", // 无需解冻
			expectedError:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currency, amount, err := logic.calculateUnfreezeAmount(tt.order, tradingPair)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCurrency, currency)
				assert.Equal(t, tt.expectedAmount, amount)
			}
		})
	}
}

func TestCancelOrderLogic_Integration(t *testing.T) {
	// 集成测试：测试完整的订单取消流程
	mockOrderModel := &mockOrderModel{}
	mockTradingPairModel := &mockTradingPairModel{}
	mockBalanceModel := &mockBalanceModel{}

	// 测试数据
	userID := uint64(1)
	orderID := uint64(123)
	order := &model.Order{
		ID:           orderID,
		UserID:       userID,
		Symbol:       "BTC/USDT",
		Type:         1, // 限价单
		Side:         1, // 买入
		Amount:       "2.00000000",
		Price:        "50000.00000000",
		FilledAmount: "0.50000000", // 部分成交
		Status:       2,            // 部分成交状态
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tradingPair := &model.TradingPair{
		ID:            1,
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
		Status:        1,
	}

	// 设置mock期望
	mockOrderModel.On("FindOne", mock.Anything, orderID).Return(order, nil)
	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)
	mockOrderModel.On("UpdateStatus", mock.Anything, orderID, int64(4)).Return(nil)
	mockOrderModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(nil)

	// 预期解冻：(2.0 - 0.5) * 50000 = 75000 USDT
	mockBalanceModel.On("UnfreezeBalance", mock.Anything, userID, "USDT", "75000").Return(nil)

	// 创建服务上下文
	svcCtx := &svc.ServiceContext{
		OrderModel:       mockOrderModel,
		TradingPairModel: mockTradingPairModel,
		BalanceModel:     mockBalanceModel,
	}

	// 创建带用户ID的上下文
	ctx := context.WithValue(context.Background(), "userId", strconv.FormatUint(userID, 10))

	// 创建逻辑实例
	logic := NewCancelOrderLogic(ctx, svcCtx)

	// 执行测试
	req := &types.CancelOrderRequest{OrderID: orderID}
	resp, err := logic.CancelOrder(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "Order canceled successfully", resp.Message)

	// 验证所有mock调用
	mockOrderModel.AssertExpectations(t)
	mockTradingPairModel.AssertExpectations(t)
	mockBalanceModel.AssertExpectations(t)
}