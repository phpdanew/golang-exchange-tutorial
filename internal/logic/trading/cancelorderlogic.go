package trading

import (
	"context"
	"errors"
	"strconv"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelOrderLogic) CancelOrder(req *types.CancelOrderRequest) (resp *types.BaseResponse, err error) {
	// 从JWT中获取用户ID
	userID, err := l.getUserIDFromContext()
	if err != nil {
		return nil, err
	}

	// 查询订单信息
	order, err := l.svcCtx.OrderModel.FindOne(l.ctx, req.OrderID)
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

	// 检查订单状态
	if order.Status == 4 { // 已取消
		return nil, model.ErrOrderAlreadyCanceled
	}

	if order.Status == 3 { // 完全成交
		return nil, model.ErrOrderAlreadyFilled
	}

	// 获取交易对信息
	tradingPair, err := l.svcCtx.TradingPairModel.FindBySymbol(l.ctx, order.Symbol)
	if err != nil {
		return nil, err
	}

	// 计算需要解冻的资产
	unfreezeCurrency, unfreezeAmount, err := l.calculateUnfreezeAmount(order, tradingPair)
	if err != nil {
		return nil, err
	}

	// 使用事务确保原子性
	err = l.svcCtx.OrderModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 更新订单状态为已取消
		if err := l.svcCtx.OrderModel.UpdateStatus(ctx, order.ID, 4); err != nil {
			return err
		}

		// 解冻用户余额
		if unfreezeAmount != "0" {
			if err := l.svcCtx.BalanceModel.UnfreezeBalance(ctx, userID, unfreezeCurrency, unfreezeAmount); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	resp = &types.BaseResponse{
		Code:    0,
		Message: "Order canceled successfully",
	}

	return resp, nil
}

// getUserIDFromContext 从上下文中获取用户ID
func (l *CancelOrderLogic) getUserIDFromContext() (uint64, error) {
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

// calculateUnfreezeAmount 计算需要解冻的资产数量
func (l *CancelOrderLogic) calculateUnfreezeAmount(order *model.Order, tradingPair *model.TradingPair) (currency string, amount string, err error) {
	// 计算剩余未成交数量
	orderAmount, err := decimal.NewFromString(order.Amount)
	if err != nil {
		return "", "", model.ErrInvalidAmount
	}

	filledAmount, err := decimal.NewFromString(order.FilledAmount)
	if err != nil {
		return "", "", model.ErrInvalidAmount
	}

	remainingAmount := orderAmount.Sub(filledAmount)

	// 如果没有剩余数量，则不需要解冻
	if remainingAmount.LessThanOrEqual(decimal.Zero) {
		return "", "0", nil
	}

	if order.Side == 1 { // 买入订单
		// 买入订单解冻计价币种（如USDT）
		currency = tradingPair.QuoteCurrency

		if order.Type == 1 { // 限价买单
			// 解冻金额 = 剩余数量 * 价格
			price, err := decimal.NewFromString(order.Price)
			if err != nil {
				return "", "", errors.New("invalid price format")
			}
			unfreezeAmount := remainingAmount.Mul(price)
			return currency, unfreezeAmount.String(), nil
		} else { // 市价买单
			// 市价买单解冻剩余的计价币种数量
			return currency, remainingAmount.String(), nil
		}
	} else { // 卖出订单
		// 卖出订单解冻基础币种（如BTC）
		currency = tradingPair.BaseCurrency
		// 解冻数量 = 剩余数量
		return currency, remainingAmount.String(), nil
	}
}
