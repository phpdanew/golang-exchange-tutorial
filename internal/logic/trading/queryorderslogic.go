package trading

import (
	"context"
	"strconv"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOrdersLogic {
	return &QueryOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryOrdersLogic) QueryOrders(req *types.QueryOrdersRequest) (resp *types.OrderListResponse, err error) {
	// 从JWT中获取用户ID
	userID, err := l.getUserIDFromContext()
	if err != nil {
		return nil, err
	}

	// 设置默认分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}

	size := req.Size
	if size <= 0 {
		size = 20
	}

	// 调用分页查询方法
	orders, total, err := l.svcCtx.OrderModel.FindByUserIDWithPagination(
		l.ctx, userID, req.Symbol, req.Status, page, size)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	var orderList []types.Order
	for _, order := range orders {
		orderList = append(orderList, types.Order{
			ID:           order.ID,
			UserID:       order.UserID,
			Symbol:       order.Symbol,
			Type:         order.Type,
			Side:         order.Side,
			Amount:       order.Amount,
			Price:        order.Price,
			FilledAmount: order.FilledAmount,
			Status:       order.Status,
			CreatedAt:    order.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    order.UpdatedAt.Format(time.RFC3339),
		})
	}

	resp = &types.OrderListResponse{
		Orders: orderList,
		Total:  total,
		Page:   page,
		Size:   size,
	}

	return resp, nil
}

// getUserIDFromContext 从上下文中获取用户ID
func (l *QueryOrdersLogic) getUserIDFromContext() (uint64, error) {
	// 这里应该从JWT中间件设置的上下文中获取用户ID
	// 为了演示，这里返回一个固定值，实际应用中需要从JWT token中解析
	userIDStr := l.ctx.Value("userId")
	if userIDStr == nil {
		return 0, model.ErrUnauthorized
	}

	userID, err := strconv.ParseUint(userIDStr.(string), 10, 64)
	if err != nil {
		return 0, model.ErrUnauthorized
	}

	return userID, nil
}
