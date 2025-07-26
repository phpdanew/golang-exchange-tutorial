package trading

import (
	"context"
	"errors"
	"testing"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/logx"
)

func TestGetOrderLogic_GetOrder_Success(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据
	now := time.Now()
	order := &model.Order{
		ID:           123,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1, // 限价单
		Side:         1, // 买入
		Amount:       "1.00000000",
		Price:        "50000.00",
		FilledAmount: "0.30000000",
		Status:       2, // 部分成交
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 设置mock预期
	mockOrderModel.On("FindOne", mock.Anything, uint64(123)).Return(order, nil)

	// 执行测试
	resp, err := logic.GetOrder("123")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(123), resp.ID)
	assert.Equal(t, uint64(1), resp.UserID)
	assert.Equal(t, "BTC/USDT", resp.Symbol)
	assert.Equal(t, int64(1), resp.Type)
	assert.Equal(t, int64(1), resp.Side)
	assert.Equal(t, "1.00000000", resp.Amount)
	assert.Equal(t, "50000.00", resp.Price)
	assert.Equal(t, "0.30000000", resp.FilledAmount)
	assert.Equal(t, int64(2), resp.Status)
	assert.Equal(t, now.Format(time.RFC3339), resp.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), resp.UpdatedAt)

	mockOrderModel.AssertExpectations(t)
}

func TestGetOrderLogic_GetOrder_OrderNotFound(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 设置mock预期 - 订单不存在
	mockOrderModel.On("FindOne", mock.Anything, uint64(999)).Return(nil, model.ErrNotFound)

	// 执行测试
	resp, err := logic.GetOrder("999")

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrOrderNotFound, err)

	mockOrderModel.AssertExpectations(t)
}

func TestGetOrderLogic_GetOrder_InvalidOrderID(t *testing.T) {
	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{}

	logic := &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 执行测试 - 无效的订单ID格式
	resp, err := logic.GetOrder("invalid_id")

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrInvalidParams, err)
}

func TestGetOrderLogic_GetOrder_NotOwner(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "2") // 当前用户ID是2
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 订单属于用户1
	order := &model.Order{
		ID:     123,
		UserID: 1, // 订单属于用户1，但当前用户是2
		Symbol: "BTC/USDT",
		Type:   1,
		Side:   1,
		Amount: "1.00000000",
		Price:  "50000.00",
		Status: 1,
	}

	mockOrderModel.On("FindOne", mock.Anything, uint64(123)).Return(order, nil)

	// 执行测试
	resp, err := logic.GetOrder("123")

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrForbidden, err)

	mockOrderModel.AssertExpectations(t)
}

func TestGetOrderLogic_GetOrder_Unauthorized(t *testing.T) {
	ctx := context.Background() // 没有设置userId

	svcCtx := &svc.ServiceContext{}

	logic := &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 执行测试
	resp, err := logic.GetOrder("123")

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrUnauthorized, err)
}

func TestGetOrderLogic_GetOrder_DatabaseError(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 设置mock预期 - 数据库错误
	dbError := errors.New("database connection failed")
	mockOrderModel.On("FindOne", mock.Anything, uint64(123)).Return(nil, dbError)

	// 执行测试
	resp, err := logic.GetOrder("123")

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "database connection failed")

	mockOrderModel.AssertExpectations(t)
}

func TestGetOrderLogic_GetOrder_MarketOrder(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 市价单
	now := time.Now()
	order := &model.Order{
		ID:           456,
		UserID:       1,
		Symbol:       "ETH/USDT",
		Type:         2, // 市价单
		Side:         2, // 卖出
		Amount:       "10.00000000",
		Price:        "", // 市价单无价格
		FilledAmount: "10.00000000",
		Status:       3, // 完全成交
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mockOrderModel.On("FindOne", mock.Anything, uint64(456)).Return(order, nil)

	// 执行测试
	resp, err := logic.GetOrder("456")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(456), resp.ID)
	assert.Equal(t, "ETH/USDT", resp.Symbol)
	assert.Equal(t, int64(2), resp.Type) // 市价单
	assert.Equal(t, int64(2), resp.Side) // 卖出
	assert.Equal(t, "", resp.Price)      // 市价单无价格
	assert.Equal(t, int64(3), resp.Status) // 完全成交

	mockOrderModel.AssertExpectations(t)
}

func TestGetOrderLogic_GetOrder_CanceledOrder(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 已取消的订单
	now := time.Now()
	order := &model.Order{
		ID:           789,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1,
		Side:         1,
		Amount:       "0.50000000",
		Price:        "48000.00",
		FilledAmount: "0.00000000", // 未成交就被取消
		Status:       4,            // 已取消
		CreatedAt:    now.Add(-time.Hour),
		UpdatedAt:    now,
	}

	mockOrderModel.On("FindOne", mock.Anything, uint64(789)).Return(order, nil)

	// 执行测试
	resp, err := logic.GetOrder("789")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(789), resp.ID)
	assert.Equal(t, "0.00000000", resp.FilledAmount) // 未成交
	assert.Equal(t, int64(4), resp.Status)           // 已取消

	mockOrderModel.AssertExpectations(t)
}

func TestGetOrderLogic_GetOrder_PendingOrder(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 待成交订单
	now := time.Now()
	order := &model.Order{
		ID:           111,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1,
		Side:         2, // 卖出
		Amount:       "2.00000000",
		Price:        "52000.00",
		FilledAmount: "0.00000000", // 未成交
		Status:       1,            // 待成交
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mockOrderModel.On("FindOne", mock.Anything, uint64(111)).Return(order, nil)

	// 执行测试
	resp, err := logic.GetOrder("111")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(111), resp.ID)
	assert.Equal(t, int64(2), resp.Side)           // 卖出
	assert.Equal(t, "0.00000000", resp.FilledAmount) // 未成交
	assert.Equal(t, int64(1), resp.Status)         // 待成交

	mockOrderModel.AssertExpectations(t)
} 