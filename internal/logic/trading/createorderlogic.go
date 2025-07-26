package trading

import (
	"context"
	"errors"
	"strconv"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderRequest) (resp *types.Order, err error) {
	// 从JWT中获取用户ID（这里假设已经通过中间件设置）
	userID, err := l.getUserIDFromContext()
	if err != nil {
		return nil, err
	}

	// 验证交易对是否存在且可用
	tradingPair, err := l.svcCtx.TradingPairModel.FindBySymbol(l.ctx, req.Symbol)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, model.ErrTradingPairNotFound
		}
		return nil, err
	}

	if tradingPair.Status != 1 {
		return nil, model.ErrTradingPairDisabled
	}

	// 验证订单参数
	if err := l.validateOrderRequest(req, tradingPair); err != nil {
		return nil, err
	}

	// 计算需要冻结的资产
	freezeCurrency, freezeAmount, err := l.calculateFreezeAmount(req, tradingPair)
	if err != nil {
		return nil, err
	}

	// 使用事务确保原子性
	var order *model.Order
	err = l.svcCtx.BalanceModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 冻结用户余额
		if err := l.svcCtx.BalanceModel.FreezeBalance(ctx, userID, freezeCurrency, freezeAmount); err != nil {
			return err
		}

		// 创建订单
		now := time.Now()
		order = &model.Order{
			UserID:       userID,
			Symbol:       req.Symbol,
			Type:         req.Type,
			Side:         req.Side,
			Amount:       req.Amount,
			Price:        req.Price,
			FilledAmount: "0",
			Status:       1, // 待成交
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		result, err := l.svcCtx.OrderModel.Insert(ctx, order)
		if err != nil {
			return err
		}

		// 获取插入的订单ID
		orderID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		order.ID = uint64(orderID)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 创建订单成功后，调用撮合引擎进行撮合处理
	matchingService := NewMatchingService(l.ctx, l.svcCtx)
	if err := matchingService.ProcessOrderWithMatching(order); err != nil {
		l.Errorf("Failed to process order with matching engine: %v", err)
		// 撮合失败不影响订单创建，只记录错误日志
		// 在生产环境中可能需要更精细的错误处理策略
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

	l.Infof("Order created successfully: ID=%d, Symbol=%s, Type=%d, Side=%d, Amount=%s, Price=%s", 
		order.ID, order.Symbol, order.Type, order.Side, order.Amount, order.Price)

	return resp, nil
}

// getUserIDFromContext 从上下文中获取用户ID
func (l *CreateOrderLogic) getUserIDFromContext() (uint64, error) {
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

// validateOrderRequest 验证订单请求参数
func (l *CreateOrderLogic) validateOrderRequest(req *types.CreateOrderRequest, tradingPair *model.TradingPair) error {
	// 验证订单类型
	if req.Type != 1 && req.Type != 2 {
		return model.ErrInvalidOrderType
	}

	// 验证交易方向
	if req.Side != 1 && req.Side != 2 {
		return model.ErrInvalidOrderSide
	}

	// 验证数量格式和精度
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return model.ErrInvalidAmount
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return model.ErrInvalidAmount
	}

	// 检查数量精度
	if amount.Exponent() < -int32(tradingPair.AmountScale) {
		return errors.New("amount precision exceeds allowed scale")
	}

	// 检查最小/最大数量限制
	minAmount, _ := decimal.NewFromString(tradingPair.MinAmount)
	maxAmount, _ := decimal.NewFromString(tradingPair.MaxAmount)

	if amount.LessThan(minAmount) {
		return errors.New("amount below minimum limit")
	}

	if maxAmount.GreaterThan(decimal.Zero) && amount.GreaterThan(maxAmount) {
		return errors.New("amount exceeds maximum limit")
	}

	// 限价单需要验证价格
	if req.Type == 1 {
		if req.Price == "" {
			return errors.New("price is required for limit orders")
		}

		price, err := decimal.NewFromString(req.Price)
		if err != nil {
			return errors.New("invalid price format")
		}

		if price.LessThanOrEqual(decimal.Zero) {
			return errors.New("price must be positive")
		}

		// 检查价格精度
		if price.Exponent() < -int32(tradingPair.PriceScale) {
			return errors.New("price precision exceeds allowed scale")
		}
	}

	return nil
}

// calculateFreezeAmount 计算需要冻结的资产数量
func (l *CreateOrderLogic) calculateFreezeAmount(req *types.CreateOrderRequest, tradingPair *model.TradingPair) (currency string, amount string, err error) {
	orderAmount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return "", "", model.ErrInvalidAmount
	}

	if req.Side == 1 { // 买入订单
		// 买入需要冻结计价币种（如USDT）
		currency = tradingPair.QuoteCurrency

		if req.Type == 1 { // 限价买单
			// 冻结金额 = 数量 * 价格
			price, err := decimal.NewFromString(req.Price)
			if err != nil {
				return "", "", errors.New("invalid price format")
			}
			freezeAmount := orderAmount.Mul(price)
			return currency, freezeAmount.String(), nil
		} else { // 市价买单
			// 市价买单的Amount字段表示要花费的计价币种数量
			return currency, req.Amount, nil
		}
	} else { // 卖出订单
		// 卖出需要冻结基础币种（如BTC）
		currency = tradingPair.BaseCurrency
		// 冻结数量 = 订单数量
		return currency, orderAmount.String(), nil
	}
}
