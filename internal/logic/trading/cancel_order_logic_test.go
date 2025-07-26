package trading

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

func TestCancelOrderLogic_CancelOrder_Success(t *testing.T) {
	mockOrderModel := &mockOrderModel{}
	mockTradingPairModel := &mockTradingPairModel{}
	mockBalanceModel := &mockBalanceModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel:       mockOrderModel,
		TradingPairModel: mockTradingPairModel,
		BalanceModel:     mockBalanceModel,
	}

	logic := &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据
	now := time.Now()
	order := &model.Order{
		ID:           1,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1, // 限价单
		Side:         1, // 买入
		Amount:       "1.00000000",
		Price:        "50000.00",
		FilledAmount: "0.30000000", // 部分成交
		Status:       2,            // 部分成交状态
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	tradingPair := &model.TradingPair{
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
	}

	// 设置mock预期
	mockOrderModel.On("FindOne", mock.Anything, uint64(1)).Return(order, nil)
	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)
	mockOrderModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(nil)
	mockOrderModel.On("UpdateStatus", mock.Anything, uint64(1), int64(4)).Return(nil)
	// 剩余未成交: 1.0 - 0.3 = 0.7, 需要解冻: 0.7 * 50000 = 35000 USDT
	mockBalanceModel.On("UnfreezeBalance", mock.Anything, uint64(1), "USDT", "35000").Return(nil)

	req := &types.CancelOrderRequest{
		OrderID: 1,
	}

	// 执行测试
	resp, err := logic.CancelOrder(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "Order canceled successfully", resp.Message)

	// 验证mock调用
	mockOrderModel.AssertExpectations(t)
	mockTradingPairModel.AssertExpectations(t)
	mockBalanceModel.AssertExpectations(t)
}

func TestCancelOrderLogic_CancelOrder_OrderNotFound(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 设置mock预期 - 订单不存在
	mockOrderModel.On("FindOne", mock.Anything, uint64(999)).Return(nil, model.ErrNotFound)

	req := &types.CancelOrderRequest{
		OrderID: 999,
	}

	// 执行测试
	resp, err := logic.CancelOrder(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrOrderNotFound, err)

	mockOrderModel.AssertExpectations(t)
}

func TestCancelOrderLogic_CancelOrder_NotOwner(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "2") // 不同的用户ID
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 订单属于用户1，但当前用户是2
	order := &model.Order{
		ID:     1,
		UserID: 1, // 订单属于用户1
		Symbol: "BTC/USDT",
		Status: 1,
	}

	mockOrderModel.On("FindOne", mock.Anything, uint64(1)).Return(order, nil)

	req := &types.CancelOrderRequest{
		OrderID: 1,
	}

	// 执行测试
	resp, err := logic.CancelOrder(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrForbidden, err)

	mockOrderModel.AssertExpectations(t)
}

func TestCancelOrderLogic_CancelOrder_AlreadyCanceled(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 已取消的订单
	order := &model.Order{
		ID:     1,
		UserID: 1,
		Symbol: "BTC/USDT",
		Status: 4, // 已取消
	}

	mockOrderModel.On("FindOne", mock.Anything, uint64(1)).Return(order, nil)

	req := &types.CancelOrderRequest{
		OrderID: 1,
	}

	// 执行测试
	resp, err := logic.CancelOrder(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrOrderAlreadyCanceled, err)

	mockOrderModel.AssertExpectations(t)
}

func TestCancelOrderLogic_CancelOrder_AlreadyFilled(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 完全成交的订单
	order := &model.Order{
		ID:     1,
		UserID: 1,
		Symbol: "BTC/USDT",
		Status: 3, // 完全成交
	}

	mockOrderModel.On("FindOne", mock.Anything, uint64(1)).Return(order, nil)

	req := &types.CancelOrderRequest{
		OrderID: 1,
	}

	// 执行测试
	resp, err := logic.CancelOrder(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrOrderAlreadyFilled, err)

	mockOrderModel.AssertExpectations(t)
}

func TestCancelOrderLogic_CancelOrder_SellOrder(t *testing.T) {
	mockOrderModel := &mockOrderModel{}
	mockTradingPairModel := &mockTradingPairModel{}
	mockBalanceModel := &mockBalanceModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel:       mockOrderModel,
		TradingPairModel: mockTradingPairModel,
		BalanceModel:     mockBalanceModel,
	}

	logic := &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 卖出订单
	order := &model.Order{
		ID:           2,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1, // 限价单
		Side:         2, // 卖出
		Amount:       "1.00000000",
		Price:        "50000.00",
		FilledAmount: "0.20000000", // 部分成交
		Status:       2,            // 部分成交状态
	}

	tradingPair := &model.TradingPair{
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
	}

	// 设置mock预期
	mockOrderModel.On("FindOne", mock.Anything, uint64(2)).Return(order, nil)
	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)
	mockOrderModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(nil)
	mockOrderModel.On("UpdateStatus", mock.Anything, uint64(2), int64(4)).Return(nil)
	// 卖出订单剩余未成交: 1.0 - 0.2 = 0.8 BTC 需要解冻
	mockBalanceModel.On("UnfreezeBalance", mock.Anything, uint64(1), "BTC", "0.8").Return(nil)

	req := &types.CancelOrderRequest{
		OrderID: 2,
	}

	// 执行测试
	resp, err := logic.CancelOrder(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, resp.Code)

	mockOrderModel.AssertExpectations(t)
	mockTradingPairModel.AssertExpectations(t)
	mockBalanceModel.AssertExpectations(t)
}

func TestCancelOrderLogic_CancelOrder_MarketOrder(t *testing.T) {
	mockOrderModel := &mockOrderModel{}
	mockTradingPairModel := &mockTradingPairModel{}
	mockBalanceModel := &mockBalanceModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel:       mockOrderModel,
		TradingPairModel: mockTradingPairModel,
		BalanceModel:     mockBalanceModel,
	}

	logic := &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 市价买单
	order := &model.Order{
		ID:           3,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         2, // 市价单
		Side:         1, // 买入
		Amount:       "1000.00", // 市价买单Amount表示USDT数量
		Price:        "",
		FilledAmount: "300.00", // 已花费300 USDT
		Status:       2,        // 部分成交状态
	}

	tradingPair := &model.TradingPair{
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
	}

	// 设置mock预期
	mockOrderModel.On("FindOne", mock.Anything, uint64(3)).Return(order, nil)
	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)
	mockOrderModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(nil)
	mockOrderModel.On("UpdateStatus", mock.Anything, uint64(3), int64(4)).Return(nil)
	// 市价买单剩余: 1000 - 300 = 700 USDT 需要解冻
	mockBalanceModel.On("UnfreezeBalance", mock.Anything, uint64(1), "USDT", "700").Return(nil)

	req := &types.CancelOrderRequest{
		OrderID: 3,
	}

	// 执行测试
	resp, err := logic.CancelOrder(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	mockOrderModel.AssertExpectations(t)
	mockTradingPairModel.AssertExpectations(t)
	mockBalanceModel.AssertExpectations(t)
}

func TestCancelOrderLogic_CancelOrder_FullyFilledNoUnfreeze(t *testing.T) {
	mockOrderModel := &mockOrderModel{}
	mockTradingPairModel := &mockTradingPairModel{}
	mockBalanceModel := &mockBalanceModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel:       mockOrderModel,
		TradingPairModel: mockTradingPairModel,
		BalanceModel:     mockBalanceModel,
	}

	logic := &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 虽然状态是部分成交，但实际上amount和filledAmount相等
	order := &model.Order{
		ID:           4,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1,
		Side:         1,
		Amount:       "1.00000000",
		Price:        "50000.00",
		FilledAmount: "1.00000000", // 实际上已经完全成交
		Status:       2,            // 但状态还是部分成交（边界情况）
	}

	tradingPair := &model.TradingPair{
		Symbol:        "BTC/USDT",
		BaseCurrency:  "BTC",
		QuoteCurrency: "USDT",
	}

	// 设置mock预期
	mockOrderModel.On("FindOne", mock.Anything, uint64(4)).Return(order, nil)
	mockTradingPairModel.On("FindBySymbol", mock.Anything, "BTC/USDT").Return(tradingPair, nil)
	mockOrderModel.On("Trans", mock.Anything, mock.AnythingOfType("func(context.Context, sqlx.Session) error")).Return(nil)
	mockOrderModel.On("UpdateStatus", mock.Anything, uint64(4), int64(4)).Return(nil)
	// 不应该调用UnfreezeBalance，因为没有剩余数量

	req := &types.CancelOrderRequest{
		OrderID: 4,
	}

	// 执行测试
	resp, err := logic.CancelOrder(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	mockOrderModel.AssertExpectations(t)
	mockTradingPairModel.AssertExpectations(t)
	mockBalanceModel.AssertExpectations(t)
}

func TestCancelOrderLogic_CancelOrder_Unauthorized(t *testing.T) {
	ctx := context.Background() // 没有设置userId

	svcCtx := &svc.ServiceContext{}

	logic := &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	req := &types.CancelOrderRequest{
		OrderID: 1,
	}

	// 执行测试
	resp, err := logic.CancelOrder(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrUnauthorized, err)
} 