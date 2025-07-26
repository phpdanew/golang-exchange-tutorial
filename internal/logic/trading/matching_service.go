package trading

import (
	"context"
	"errors"
	"strconv"

	"crypto-exchange/internal/matching"
	"crypto-exchange/internal/svc"
	"crypto-exchange/model"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// MatchingService 撮合服务，负责协调撮合引擎和成交执行
type MatchingService struct {
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	logger  logx.Logger
}

// NewMatchingService 创建新的撮合服务
func NewMatchingService(ctx context.Context, svcCtx *svc.ServiceContext) *MatchingService {
	return &MatchingService{
		ctx:    ctx,
		svcCtx: svcCtx,
		logger: logx.WithContext(ctx),
	}
}

// ProcessOrderWithMatching 处理订单并执行撮合
func (ms *MatchingService) ProcessOrderWithMatching(order *model.Order) error {
	// 1. 使用撮合引擎处理订单
	matchResult, err := ms.svcCtx.MatchingEngine.ProcessOrder(order)
	if err != nil {
		ms.logger.Errorf("Failed to process order in matching engine: %v", err)
		return err
	}

	// 2. 如果有成交，执行成交记录和余额更新
	if len(matchResult.Trades) > 0 {
		if err := ms.executeTrades(matchResult); err != nil {
			ms.logger.Errorf("Failed to execute trades: %v", err)
			return err
		}
	}

	// 3. 更新订单状态
	if err := ms.updateOrderStatus(order); err != nil {
		ms.logger.Errorf("Failed to update order status: %v", err)
		return err
	}

	ms.logger.Infof("Order %d processed successfully with %d trades", order.ID, len(matchResult.Trades))
	return nil
}

// executeTrades 执行成交记录列表，确保事务一致性
func (ms *MatchingService) executeTrades(matchResult *matching.MatchResult) error {
	if len(matchResult.Trades) == 0 {
		return nil
	}

	// 使用数据库事务确保所有操作的原子性
	return ms.svcCtx.BalanceModel.Trans(ms.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 创建所有成交记录
		for _, trade := range matchResult.Trades {
			if err := ms.createTradeRecord(ctx, trade); err != nil {
				ms.logger.Errorf("Failed to create trade record: %v", err)
				return err
			}
		}

		// 2. 更新所有订单状态
		for _, order := range matchResult.UpdatedOrders {
			if err := ms.updateOrderInDB(ctx, order); err != nil {
				ms.logger.Errorf("Failed to update order status: %v", err)
				return err
			}
		}

		// 3. 处理余额更新
		if err := ms.updateUserBalances(ctx, matchResult.Trades); err != nil {
			ms.logger.Errorf("Failed to update user balances: %v", err)
			return err
		}

		// 4. 解冻已完全成交订单的余额
		if err := ms.unfreezeCompletedOrderBalances(ctx, matchResult.FilledOrders); err != nil {
			ms.logger.Errorf("Failed to unfreeze balances: %v", err)
			return err
		}

		ms.logger.Infof("Successfully executed %d trades", len(matchResult.Trades))
		return nil
	})
}

// createTradeRecord 创建成交记录
func (ms *MatchingService) createTradeRecord(ctx context.Context, trade *model.Trade) error {
	_, err := ms.svcCtx.TradeModel.Insert(ctx, trade)
	if err != nil {
		return err
	}

	ms.logger.Infof("Trade record created: %s %s@%s between user %d and %d",
		trade.Symbol, trade.Amount, trade.Price, trade.BuyUserID, trade.SellUserID)
	return nil
}

// updateOrderInDB 更新订单状态到数据库
func (ms *MatchingService) updateOrderInDB(ctx context.Context, order *model.Order) error {
	err := ms.svcCtx.OrderModel.Update(ctx, order)
	if err != nil {
		return err
	}

	ms.logger.Infof("Order %d status updated to %d, filled amount: %s",
		order.ID, order.Status, order.FilledAmount)
	return nil
}

// updateOrderStatus 更新单个订单状态
func (ms *MatchingService) updateOrderStatus(order *model.Order) error {
	return ms.svcCtx.OrderModel.Update(ms.ctx, order)
}

// updateUserBalances 更新用户余额
func (ms *MatchingService) updateUserBalances(ctx context.Context, trades []*model.Trade) error {
	// 聚合同一用户的余额变动
	balanceChanges := make(map[string]decimal.Decimal) // userID_currency -> amount

	for _, trade := range trades {
		tradePrice, _ := decimal.NewFromString(trade.Price)
		tradeAmount, _ := decimal.NewFromString(trade.Amount)
		totalValue := tradePrice.Mul(tradeAmount)

		// 从交易对符号中提取基础币种和计价币种
		baseCurrency, quoteCurrency, err := ms.parseSymbol(trade.Symbol)
		if err != nil {
			return err
		}

		// 买方：减少计价币种，增加基础币种
		buyerQuoteKey := ms.getUserCurrencyKey(trade.BuyUserID, quoteCurrency)
		buyerBaseKey := ms.getUserCurrencyKey(trade.BuyUserID, baseCurrency)
		
		balanceChanges[buyerQuoteKey] = balanceChanges[buyerQuoteKey].Sub(totalValue)
		balanceChanges[buyerBaseKey] = balanceChanges[buyerBaseKey].Add(tradeAmount)

		// 卖方：减少基础币种，增加计价币种
		sellerBaseKey := ms.getUserCurrencyKey(trade.SellUserID, baseCurrency)
		sellerQuoteKey := ms.getUserCurrencyKey(trade.SellUserID, quoteCurrency)
		
		balanceChanges[sellerBaseKey] = balanceChanges[sellerBaseKey].Sub(tradeAmount)
		balanceChanges[sellerQuoteKey] = balanceChanges[sellerQuoteKey].Add(totalValue)
	}

	// 应用余额变动
	for userCurrencyKey, amount := range balanceChanges {
		userID, currency := ms.parseUserCurrencyKey(userCurrencyKey)
		if err := ms.applyBalanceChange(ctx, userID, currency, amount); err != nil {
			return err
		}
	}

	return nil
}

// unfreezeCompletedOrderBalances 解冻已完全成交订单的余额
func (ms *MatchingService) unfreezeCompletedOrderBalances(ctx context.Context, filledOrders []*model.Order) error {
	for _, order := range filledOrders {
		if order.Status != 3 { // 只处理完全成交的订单
			continue
		}

		// 从交易对符号中提取基础币种和计价币种
		baseCurrency, quoteCurrency, err := ms.parseSymbol(order.Symbol)
		if err != nil {
			return err
		}

		orderAmount, _ := decimal.NewFromString(order.Amount)
		
		var currency string
		var unfreezeAmount decimal.Decimal

		if order.Side == 1 { // 买单，解冻计价币种
			currency = quoteCurrency
			orderPrice, _ := decimal.NewFromString(order.Price)
			unfreezeAmount = orderAmount.Mul(orderPrice)
		} else { // 卖单，解冻基础币种
			currency = baseCurrency
			unfreezeAmount = orderAmount
		}

		// 执行解冻操作
		err = ms.svcCtx.BalanceModel.UnfreezeBalance(ctx, order.UserID, currency, unfreezeAmount.String())
		if err != nil {
			ms.logger.Errorf("Failed to unfreeze balance for order %d: %v", order.ID, err)
			return err
		}

		ms.logger.Infof("Unfroze %s %s for completed order %d", unfreezeAmount.String(), currency, order.ID)
	}

	return nil
}

// applyBalanceChange 应用余额变动
func (ms *MatchingService) applyBalanceChange(ctx context.Context, userID uint64, currency string, amount decimal.Decimal) error {
	if amount.IsZero() {
		return nil
	}

	// 查找用户余额记录
	balance, err := ms.svcCtx.BalanceModel.FindByUserIDAndCurrency(ctx, userID, currency)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 余额记录不存在，创建新记录
			newBalance := &model.Balance{
				UserID:    userID,
				Currency:  currency,
				Available: amount.String(),
				Frozen:    "0",
			}
			_, err = ms.svcCtx.BalanceModel.Insert(ctx, newBalance)
			return err
		}
		return err
	}

	// 更新现有余额
	currentAvailable, _ := decimal.NewFromString(balance.Available)
	newAvailable := currentAvailable.Add(amount)
	
	if newAvailable.LessThan(decimal.Zero) {
		return errors.New("insufficient balance after trade execution")
	}

	return ms.svcCtx.BalanceModel.UpdateBalance(ctx, userID, currency, newAvailable.String(), balance.Frozen)
}

// parseSymbol 解析交易对符号，返回基础币种和计价币种
func (ms *MatchingService) parseSymbol(symbol string) (string, string, error) {
	// 查找 "/" 分隔符
	for i, char := range symbol {
		if char == '/' {
			if i == 0 || i == len(symbol)-1 {
				return "", "", errors.New("invalid symbol format")
			}
			return symbol[:i], symbol[i+1:], nil
		}
	}
	return "", "", errors.New("invalid symbol format: no separator found")
}

// getUserCurrencyKey 生成用户-币种键
func (ms *MatchingService) getUserCurrencyKey(userID uint64, currency string) string {
	return strconv.FormatUint(userID, 10) + "_" + currency
}

// parseUserCurrencyKey 解析用户-币种键
func (ms *MatchingService) parseUserCurrencyKey(key string) (uint64, string) {
	// 查找 "_" 分隔符
	for i, char := range key {
		if char == '_' {
			if i == 0 || i == len(key)-1 {
				return 0, ""
			}
			userID, err := strconv.ParseUint(key[:i], 10, 64)
			if err != nil {
				return 0, ""
			}
			return userID, key[i+1:]
		}
	}
	return 0, ""
} 