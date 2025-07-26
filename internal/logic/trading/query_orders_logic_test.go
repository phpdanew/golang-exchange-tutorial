package trading

import (
	"context"
	"errors"
	"testing"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zeromicro/go-zero/core/logx"
)

func TestQueryOrdersLogic_QueryOrders_Success(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &QueryOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据
	now := time.Now()
	orders := []*model.Order{
		{
			ID:           1,
			UserID:       1,
			Symbol:       "BTC/USDT",
			Type:         1,
			Side:         1,
			Amount:       "1.00000000",
			Price:        "50000.00",
			FilledAmount: "0.50000000",
			Status:       2, // 部分成交
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           2,
			UserID:       1,
			Symbol:       "ETH/USDT",
			Type:         2,
			Side:         2,
			Amount:       "10.00000000",
			Price:        "",
			FilledAmount: "10.00000000",
			Status:       3, // 完全成交
			CreatedAt:    now.Add(-time.Hour),
			UpdatedAt:    now.Add(-time.Hour),
		},
	}

	// 设置mock预期
	mockOrderModel.On("FindByUserIDWithPagination", mock.Anything, uint64(1), "", int64(0), int64(1), int64(20)).Return(orders, int64(2), nil)

	req := &types.QueryOrdersRequest{
		Page: 1,
		Size: 20,
	}

	// 执行测试
	resp, err := logic.QueryOrders(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(2), resp.Total)
	assert.Equal(t, int64(1), resp.Page)
	assert.Equal(t, int64(20), resp.Size)
	assert.Len(t, resp.Orders, 2)

	// 验证第一个订单
	order1 := resp.Orders[0]
	assert.Equal(t, uint64(1), order1.ID)
	assert.Equal(t, "BTC/USDT", order1.Symbol)
	assert.Equal(t, int64(1), order1.Type)
	assert.Equal(t, int64(1), order1.Side)
	assert.Equal(t, "1.00000000", order1.Amount)
	assert.Equal(t, "50000.00", order1.Price)
	assert.Equal(t, "0.50000000", order1.FilledAmount)
	assert.Equal(t, int64(2), order1.Status)

	mockOrderModel.AssertExpectations(t)
}

func TestQueryOrdersLogic_QueryOrders_WithFilters(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &QueryOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 准备测试数据 - 只返回符合过滤条件的订单
	now := time.Now()
	orders := []*model.Order{
		{
			ID:           1,
			UserID:       1,
			Symbol:       "BTC/USDT",
			Type:         1,
			Side:         1,
			Amount:       "1.00000000",
			Price:        "50000.00",
			FilledAmount: "0",
			Status:       1, // 待成交
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	// 设置mock预期 - 带过滤条件
	mockOrderModel.On("FindByUserIDWithPagination", mock.Anything, uint64(1), "BTC/USDT", int64(1), int64(1), int64(10)).Return(orders, int64(1), nil)

	req := &types.QueryOrdersRequest{
		Symbol: "BTC/USDT",
		Status: 1, // 待成交
		Page:   1,
		Size:   10,
	}

	// 执行测试
	resp, err := logic.QueryOrders(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.Total)
	assert.Equal(t, int64(1), resp.Page)
	assert.Equal(t, int64(10), resp.Size)
	assert.Len(t, resp.Orders, 1)

	// 验证订单符合过滤条件
	order := resp.Orders[0]
	assert.Equal(t, "BTC/USDT", order.Symbol)
	assert.Equal(t, int64(1), order.Status)

	mockOrderModel.AssertExpectations(t)
}

func TestQueryOrdersLogic_QueryOrders_DefaultPagination(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &QueryOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 设置mock预期 - 验证默认分页参数
	mockOrderModel.On("FindByUserIDWithPagination", mock.Anything, uint64(1), "", int64(0), int64(1), int64(20)).Return([]*model.Order{}, int64(0), nil)

	req := &types.QueryOrdersRequest{
		// 不设置Page和Size，测试默认值
	}

	// 执行测试
	resp, err := logic.QueryOrders(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.Page)  // 默认页码为1
	assert.Equal(t, int64(20), resp.Size) // 默认页面大小为20

	mockOrderModel.AssertExpectations(t)
}

func TestQueryOrdersLogic_QueryOrders_EmptyResult(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &QueryOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 设置mock预期 - 无订单
	mockOrderModel.On("FindByUserIDWithPagination", mock.Anything, uint64(1), "", int64(0), int64(1), int64(20)).Return([]*model.Order{}, int64(0), nil)

	req := &types.QueryOrdersRequest{
		Page: 1,
		Size: 20,
	}

	// 执行测试
	resp, err := logic.QueryOrders(req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(0), resp.Total)
	assert.Len(t, resp.Orders, 0)

	mockOrderModel.AssertExpectations(t)
}

func TestQueryOrdersLogic_QueryOrders_Unauthorized(t *testing.T) {
	ctx := context.Background() // 没有设置userId

	svcCtx := &svc.ServiceContext{}

	logic := &QueryOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	req := &types.QueryOrdersRequest{
		Page: 1,
		Size: 20,
	}

	// 执行测试
	resp, err := logic.QueryOrders(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, model.ErrUnauthorized, err)
}

func TestQueryOrdersLogic_QueryOrders_DatabaseError(t *testing.T) {
	mockOrderModel := &mockOrderModel{}

	ctx := context.WithValue(context.Background(), "userId", "1")
	svcCtx := &svc.ServiceContext{
		OrderModel: mockOrderModel,
	}

	logic := &QueryOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}

	// 设置mock预期 - 数据库错误
	dbError := errors.New("database connection failed")
	mockOrderModel.On("FindByUserIDWithPagination", mock.Anything, uint64(1), "", int64(0), int64(1), int64(20)).Return([]*model.Order{}, int64(0), dbError)

	req := &types.QueryOrdersRequest{
		Page: 1,
		Size: 20,
	}

	// 执行测试
	resp, err := logic.QueryOrders(req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "database connection failed")

	mockOrderModel.AssertExpectations(t)
} 