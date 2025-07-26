package matching

import (
	"context"
	"errors"
	"sync"
	"time"

	"crypto-exchange/model"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

// MatchResult 撮合结果
type MatchResult struct {
	Trades        []*model.Trade // 生成的成交记录
	UpdatedOrders []*model.Order // 更新的订单
	FilledOrders  []*model.Order // 完全成交的订单
}

// Engine 撮合引擎接口
type Engine interface {
	ProcessOrder(order *model.Order) (*MatchResult, error)
	CancelOrder(order *model.Order) error
	GetMarketDepth(symbol string, depth int) ([]PriceLevel, []PriceLevel)
	GetOrderBookSnapshot(symbol string) map[string]interface{}
}

// MatchingEngine 撮合引擎实现
type MatchingEngine struct {
	orderBooks map[string]*OrderBook // 交易对符号 -> 订单簿
	mutex      sync.RWMutex          // 读写锁，保证并发安全
	logger     logx.Logger
}

// NewMatchingEngine 创建新的撮合引擎
func NewMatchingEngine() *MatchingEngine {
	return &MatchingEngine{
		orderBooks: make(map[string]*OrderBook),
		logger:     logx.WithContext(context.Background()),
	}
}

// GetOrderBook 获取指定交易对的订单簿
func (me *MatchingEngine) GetOrderBook(symbol string) *OrderBook {
	me.mutex.RLock()
	orderBook, exists := me.orderBooks[symbol]
	me.mutex.RUnlock()

	if !exists {
		me.mutex.Lock()
		// 双重检查，避免重复创建
		if orderBook, exists = me.orderBooks[symbol]; !exists {
			orderBook = NewOrderBook(symbol)
			me.orderBooks[symbol] = orderBook
		}
		me.mutex.Unlock()
	}

	return orderBook
}

// ProcessOrder 处理新订单，执行撮合逻辑
func (me *MatchingEngine) ProcessOrder(order *model.Order) (*MatchResult, error) {
	if order == nil {
		return nil, errors.New("order cannot be nil")
	}

	orderBook := me.GetOrderBook(order.Symbol)
	result := &MatchResult{
		Trades:        make([]*model.Trade, 0),
		UpdatedOrders: make([]*model.Order, 0),
		FilledOrders:  make([]*model.Order, 0),
	}

	// 根据订单类型执行不同的撮合逻辑
	switch order.Type {
	case 1: // 限价单
		me.processLimitOrder(order, orderBook, result)
	case 2: // 市价单
		me.processMarketOrder(order, orderBook, result)
	default:
		return nil, errors.New("unsupported order type")
	}

	me.logger.Infof("Order %d processed, generated %d trades", order.ID, len(result.Trades))
	return result, nil
}

// processLimitOrder 处理限价单
func (me *MatchingEngine) processLimitOrder(order *model.Order, orderBook *OrderBook, result *MatchResult) {
	orderPrice, _ := decimal.NewFromString(order.Price)
	orderAmount, _ := decimal.NewFromString(order.Amount)
	remainingAmount := orderAmount

	if order.Side == 1 { // 买单，与卖盘撮合
		for remainingAmount.GreaterThan(decimal.Zero) {
			bestAsk, exists := orderBook.GetBestAsk()
			if !exists || bestAsk.Price.GreaterThan(orderPrice) {
				// 没有卖单或者最优卖价高于买价，停止撮合
				break
			}

			// 执行撮合
			tradeAmount, matchedOrder := me.matchOrders(order, bestAsk.Orders.Front().Value.(*model.Order), remainingAmount)
			trade := me.createTrade(order, matchedOrder, bestAsk.Price, tradeAmount)
			result.Trades = append(result.Trades, trade)

			// 更新剩余数量
			remainingAmount = remainingAmount.Sub(tradeAmount)

			// 处理被撮合的订单
			me.updateMatchedOrder(matchedOrder, tradeAmount, orderBook, result)
		}
	} else { // 卖单，与买盘撮合
		for remainingAmount.GreaterThan(decimal.Zero) {
			bestBid, exists := orderBook.GetBestBid()
			if !exists || bestBid.Price.LessThan(orderPrice) {
				// 没有买单或者最优买价低于卖价，停止撮合
				break
			}

			// 执行撮合
			tradeAmount, matchedOrder := me.matchOrders(order, bestBid.Orders.Front().Value.(*model.Order), remainingAmount)
			trade := me.createTrade(matchedOrder, order, bestBid.Price, tradeAmount)
			result.Trades = append(result.Trades, trade)

			// 更新剩余数量
			remainingAmount = remainingAmount.Sub(tradeAmount)

			// 处理被撮合的订单
			me.updateMatchedOrder(matchedOrder, tradeAmount, orderBook, result)
		}
	}

	// 更新当前订单状态
	me.updateCurrentOrder(order, remainingAmount, orderBook, result)
}

// processMarketOrder 处理市价单
func (me *MatchingEngine) processMarketOrder(order *model.Order, orderBook *OrderBook, result *MatchResult) {
	orderAmount, _ := decimal.NewFromString(order.Amount)
	remainingAmount := orderAmount

	if order.Side == 1 { // 市价买单，与卖盘撮合
		for remainingAmount.GreaterThan(decimal.Zero) {
			bestAsk, exists := orderBook.GetBestAsk()
			if !exists {
				// 没有卖单，无法成交
				break
			}

			// 执行撮合
			tradeAmount, matchedOrder := me.matchOrders(order, bestAsk.Orders.Front().Value.(*model.Order), remainingAmount)
			trade := me.createTrade(order, matchedOrder, bestAsk.Price, tradeAmount)
			result.Trades = append(result.Trades, trade)

			// 更新剩余数量
			remainingAmount = remainingAmount.Sub(tradeAmount)

			// 处理被撮合的订单
			me.updateMatchedOrder(matchedOrder, tradeAmount, orderBook, result)
		}
	} else { // 市价卖单，与买盘撮合
		for remainingAmount.GreaterThan(decimal.Zero) {
			bestBid, exists := orderBook.GetBestBid()
			if !exists {
				// 没有买单，无法成交
				break
			}

			// 执行撮合
			tradeAmount, matchedOrder := me.matchOrders(order, bestBid.Orders.Front().Value.(*model.Order), remainingAmount)
			trade := me.createTrade(matchedOrder, order, bestBid.Price, tradeAmount)
			result.Trades = append(result.Trades, trade)

			// 更新剩余数量
			remainingAmount = remainingAmount.Sub(tradeAmount)

			// 处理被撮合的订单
			me.updateMatchedOrder(matchedOrder, tradeAmount, orderBook, result)
		}
	}

	// 市价单处理完成后直接设置为完全成交或取消状态
	if remainingAmount.LessThan(orderAmount) {
		// 有部分成交
		filledAmount, _ := decimal.NewFromString(order.FilledAmount)
		order.FilledAmount = filledAmount.Add(orderAmount.Sub(remainingAmount)).String()
		
		if remainingAmount.IsZero() {
			order.Status = 3 // 完全成交
			result.FilledOrders = append(result.FilledOrders, order)
		} else {
			order.Status = 4 // 市价单未完全成交则取消
		}
		result.UpdatedOrders = append(result.UpdatedOrders, order)
	}
}

// matchOrders 撮合两个订单，返回成交数量和被撮合的订单
func (me *MatchingEngine) matchOrders(takerOrder, makerOrder *model.Order, takerRemaining decimal.Decimal) (decimal.Decimal, *model.Order) {
	makerRemaining, _ := decimal.NewFromString(makerOrder.Amount)
	makerFilled, _ := decimal.NewFromString(makerOrder.FilledAmount)
	makerAvailable := makerRemaining.Sub(makerFilled)

	// 计算成交数量（取较小值）
	tradeAmount := decimal.Min(takerRemaining, makerAvailable)

	return tradeAmount, makerOrder
}

// createTrade 创建成交记录
func (me *MatchingEngine) createTrade(buyOrder, sellOrder *model.Order, price, amount decimal.Decimal) *model.Trade {
	return &model.Trade{
		Symbol:      buyOrder.Symbol,
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		BuyUserID:   buyOrder.UserID,
		SellUserID:  sellOrder.UserID,
		Price:       price.String(),
		Amount:      amount.String(),
		CreatedAt:   time.Now(),
	}
}

// updateMatchedOrder 更新被撮合的订单
func (me *MatchingEngine) updateMatchedOrder(order *model.Order, tradeAmount decimal.Decimal, orderBook *OrderBook, result *MatchResult) {
	filledAmount, _ := decimal.NewFromString(order.FilledAmount)
	orderAmount, _ := decimal.NewFromString(order.Amount)
	
	// 更新已成交数量
	newFilledAmount := filledAmount.Add(tradeAmount)
	order.FilledAmount = newFilledAmount.String()

	// 检查是否完全成交
	if newFilledAmount.GreaterThanOrEqual(orderAmount) {
		order.Status = 3 // 完全成交
		orderBook.RemoveOrder(order)
		result.FilledOrders = append(result.FilledOrders, order)
	} else {
		order.Status = 2 // 部分成交
		// 更新订单簿中的数量
		remainingAmount := orderAmount.Sub(newFilledAmount)
		orderBook.UpdateOrderAmount(order, remainingAmount)
		result.UpdatedOrders = append(result.UpdatedOrders, order)
	}
}

// updateCurrentOrder 更新当前订单状态
func (me *MatchingEngine) updateCurrentOrder(order *model.Order, remainingAmount decimal.Decimal, orderBook *OrderBook, result *MatchResult) {
	orderAmount, _ := decimal.NewFromString(order.Amount)
	
	if remainingAmount.IsZero() {
		// 完全成交
		order.Status = 3
		order.FilledAmount = orderAmount.String()
		result.FilledOrders = append(result.FilledOrders, order)
	} else if remainingAmount.LessThan(orderAmount) {
		// 部分成交
		order.Status = 2
		filledAmount := orderAmount.Sub(remainingAmount)
		order.FilledAmount = filledAmount.String()
		order.Amount = remainingAmount.String()
		orderBook.AddOrder(order) // 将剩余部分加入订单簿
		result.UpdatedOrders = append(result.UpdatedOrders, order)
	} else {
		// 未成交
		order.Status = 1
		orderBook.AddOrder(order) // 加入订单簿等待撮合
		result.UpdatedOrders = append(result.UpdatedOrders, order)
	}
}

// CancelOrder 取消订单
func (me *MatchingEngine) CancelOrder(order *model.Order) error {
	orderBook := me.GetOrderBook(order.Symbol)
	orderBook.RemoveOrder(order)
	
	order.Status = 4 // 已取消
	me.logger.Infof("Order %d cancelled", order.ID)
	return nil
}

// GetMarketDepth 获取市场深度
func (me *MatchingEngine) GetMarketDepth(symbol string, depth int) ([]PriceLevel, []PriceLevel) {
	orderBook := me.GetOrderBook(symbol)
	return orderBook.GetDepth(depth)
}

// GetOrderBookSnapshot 获取订单簿快照（用于调试和监控）
func (me *MatchingEngine) GetOrderBookSnapshot(symbol string) map[string]interface{} {
	orderBook := me.GetOrderBook(symbol)
	orderBook.mutex.RLock()
	defer orderBook.mutex.RUnlock()

	// 统计订单簿信息
	snapshot := map[string]interface{}{
		"symbol":       symbol,
		"last_update":  orderBook.LastUpdate,
		"bid_levels":   len(orderBook.Bids),
		"ask_levels":   len(orderBook.Asks),
		"bid_prices":   len(orderBook.BidPrices),
		"ask_prices":   len(orderBook.AskPrices),
	}

	// 添加最优买卖价
	if bestBid, exists := orderBook.GetBestBid(); exists {
		snapshot["best_bid"] = bestBid.Price.String()
		snapshot["best_bid_amount"] = bestBid.Total.String()
	}

	if bestAsk, exists := orderBook.GetBestAsk(); exists {
		snapshot["best_ask"] = bestAsk.Price.String()
		snapshot["best_ask_amount"] = bestAsk.Total.String()
	}

	return snapshot
} 