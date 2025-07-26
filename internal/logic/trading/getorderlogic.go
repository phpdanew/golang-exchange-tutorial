package trading

import (
	"context"
	"errors"
	"strconv"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderLogic) GetOrder(orderID string) (resp *types.Order, err error) {
	// 从JWT中获取用户ID
	userID, err := l.getUserIDFromContext()
	if err != nil {
		return nil, err
	}

	// 转换订单ID
	id, err := strconv.ParseUint(orderID, 10, 64)
	if err != nil {
		return nil, model.ErrInvalidParams
	}

	// 查询订单信息
	order, err := l.svcCtx.OrderModel.FindOne(l.ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, model.ErrOrderNotFound
		}
		return nil, err
	}

	// 验证订单所有权
	if order.UserID != userID {
		return nil, model.ErrForbidden
	}

	// 转换为响应格式
	resp = &types.Order{
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
	}

	return resp, nil
}

// getUserIDFromContext 从上下文中获取用户ID
func (l *GetOrderLogic) getUserIDFromContext() (uint64, error) {
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
