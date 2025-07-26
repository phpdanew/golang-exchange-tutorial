package matching

import (
	"testing"
	"time"

	"crypto-exchange/model"

	"github.com/stretchr/testify/assert"
)

func TestMatchingEngine_ProcessLimitOrder_FullMatch(t *testing.T) {
	engine := NewMatchingEngine()

	// 先添加一个卖单到订单簿
	sellOrder := &model.Order{
		ID:           1,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1, // 限价单
		Side:         2, // 卖单
		Amount:       "1.0",
		Price:        "50000.0",
		FilledAmount: "0",
		Status:       1,
		CreatedAt:    time.Now(),
	}

	// 手动添加到订单簿（模拟已存在的订单）
	orderBook := engine.GetOrderBook("BTC/USDT")
	orderBook.AddOrder(sellOrder)

	// 创建匹配的买单
	buyOrder := &model.Order{
		ID:           2,
		UserID:       2,
		Symbol:       "BTC/USDT",
		Type:         1, // 限价单
		Side:         1, // 买单
		Amount:       "1.0",
		Price:        "50000.0", // 相同价格，可以成交
		FilledAmount: "0",
		Status:       1,
		CreatedAt:    time.Now(),
	}

	// 处理买单
	result, err := engine.ProcessOrder(buyOrder)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Trades))
	assert.Equal(t, 2, len(result.FilledOrders)) // 买卖双方都完全成交

	// 验证成交记录
	trade := result.Trades[0]
	assert.Equal(t, "BTC/USDT", trade.Symbol)
	assert.Equal(t, uint64(2), trade.BuyOrderID)
	assert.Equal(t, uint64(1), trade.SellOrderID)
	assert.Equal(t, uint64(2), trade.BuyUserID)
	assert.Equal(t, uint64(1), trade.SellUserID)
	assert.Equal(t, "50000", trade.Price)
	assert.Equal(t, "1", trade.Amount)

	// 验证订单状态
	assert.Equal(t, int64(3), buyOrder.Status)  // 完全成交
	assert.Equal(t, int64(3), sellOrder.Status) // 完全成交
	assert.Equal(t, "1", buyOrder.FilledAmount)
	assert.Equal(t, "1", sellOrder.FilledAmount)
}

func TestMatchingEngine_ProcessLimitOrder_PartialMatch(t *testing.T) {
	engine := NewMatchingEngine()

	// 先添加一个较小的卖单
	sellOrder := &model.Order{
		ID:           1,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1,
		Side:         2, // 卖单
		Amount:       "0.5", // 较小数量
		Price:        "50000.0",
		FilledAmount: "0",
		Status:       1,
	}

	orderBook := engine.GetOrderBook("BTC/USDT")
	orderBook.AddOrder(sellOrder)

	// 创建较大的买单
	buyOrder := &model.Order{
		ID:           2,
		UserID:       2,
		Symbol:       "BTC/USDT",
		Type:         1,
		Side:         1, // 买单
		Amount:       "1.0", // 较大数量
		Price:        "50000.0",
		FilledAmount: "0",
		Status:       1,
	}

	result, err := engine.ProcessOrder(buyOrder)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Trades))
	assert.Equal(t, 1, len(result.FilledOrders)) // 只有卖单完全成交
	assert.Equal(t, 1, len(result.UpdatedOrders)) // 买单部分成交

	// 验证成交记录
	trade := result.Trades[0]
	assert.Equal(t, "0.5", trade.Amount) // 只能成交0.5

	// 验证订单状态
	assert.Equal(t, int64(3), sellOrder.Status) // 卖单完全成交
	assert.Equal(t, int64(2), buyOrder.Status)  // 买单部分成交
	assert.Equal(t, "0.5", buyOrder.FilledAmount)
	assert.Equal(t, "0.5", sellOrder.FilledAmount)
}

func TestMatchingEngine_ProcessLimitOrder_NoMatch(t *testing.T) {
	engine := NewMatchingEngine()

	// 先添加一个高价卖单
	sellOrder := &model.Order{
		ID:     1,
		UserID: 1,
		Symbol: "BTC/USDT",
		Type:   1,
		Side:   2, // 卖单
		Amount: "1.0",
		Price:  "52000.0", // 高价
		Status: 1,
	}

	orderBook := engine.GetOrderBook("BTC/USDT")
	orderBook.AddOrder(sellOrder)

	// 创建低价买单（无法匹配）
	buyOrder := &model.Order{
		ID:           2,
		UserID:       2,
		Symbol:       "BTC/USDT",
		Type:         1,
		Side:         1, // 买单
		Amount:       "1.0",
		Price:        "50000.0", // 低价，无法匹配
		FilledAmount: "0",
		Status:       1,
	}

	result, err := engine.ProcessOrder(buyOrder)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Trades))        // 无成交
	assert.Equal(t, 0, len(result.FilledOrders))  // 无完全成交
	assert.Equal(t, 1, len(result.UpdatedOrders)) // 买单进入订单簿

	// 验证订单状态
	assert.Equal(t, int64(1), buyOrder.Status) // 待成交
	assert.Equal(t, "0", buyOrder.FilledAmount)
}

func TestMatchingEngine_ProcessMarketOrder_Buy(t *testing.T) {
	engine := NewMatchingEngine()

	// 添加卖单到订单簿
	sellOrder1 := &model.Order{
		ID:           1,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1,
		Side:         2, // 卖单
		Amount:       "0.5",
		Price:        "50000.0",
		FilledAmount: "0",
		Status:       1,
	}

	sellOrder2 := &model.Order{
		ID:           2,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1,
		Side:         2, // 卖单
		Amount:       "0.8",
		Price:        "50500.0", // 略高价格
		FilledAmount: "0",
		Status:       1,
	}

	orderBook := engine.GetOrderBook("BTC/USDT")
	orderBook.AddOrder(sellOrder1)
	orderBook.AddOrder(sellOrder2)

	// 创建市价买单
	marketBuyOrder := &model.Order{
		ID:           3,
		UserID:       2,
		Symbol:       "BTC/USDT",
		Type:         2, // 市价单
		Side:         1, // 买单
		Amount:       "1.0",
		Price:        "", // 市价单无价格
		FilledAmount: "0",
		Status:       1,
	}

	result, err := engine.ProcessOrder(marketBuyOrder)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Trades)) // 与两个卖单成交

	// 验证第一笔成交（价格优先）
	trade1 := result.Trades[0]
	assert.Equal(t, "50000", trade1.Price) // 先与低价成交
	assert.Equal(t, "0.5", trade1.Amount)

	// 验证第二笔成交
	trade2 := result.Trades[1]
	assert.Equal(t, "50500", trade2.Price) // 再与高价成交
	assert.Equal(t, "0.5", trade2.Amount)   // 剩余数量

	// 验证市价单状态
	assert.Equal(t, int64(3), marketBuyOrder.Status) // 完全成交
	assert.Equal(t, "1", marketBuyOrder.FilledAmount)
}

func TestMatchingEngine_ProcessMarketOrder_Sell(t *testing.T) {
	engine := NewMatchingEngine()

	// 添加买单到订单簿
	buyOrder1 := &model.Order{
		ID:           1,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1,
		Side:         1, // 买单
		Amount:       "0.3",
		Price:        "50000.0",
		FilledAmount: "0",
		Status:       1,
	}

	buyOrder2 := &model.Order{
		ID:           2,
		UserID:       1,
		Symbol:       "BTC/USDT",
		Type:         1,
		Side:         1, // 买单
		Amount:       "0.7",
		Price:        "49500.0", // 略低价格
		FilledAmount: "0",
		Status:       1,
	}

	orderBook := engine.GetOrderBook("BTC/USDT")
	orderBook.AddOrder(buyOrder1)
	orderBook.AddOrder(buyOrder2)

	// 创建市价卖单
	marketSellOrder := &model.Order{
		ID:           3,
		UserID:       2,
		Symbol:       "BTC/USDT",
		Type:         2, // 市价单
		Side:         2, // 卖单
		Amount:       "1.0",
		Price:        "", // 市价单无价格
		FilledAmount: "0",
		Status:       1,
	}

	result, err := engine.ProcessOrder(marketSellOrder)

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Trades)) // 与两个买单成交

	// 验证第一笔成交（价格优先，买盘从高到低）
	trade1 := result.Trades[0]
	assert.Equal(t, "50000", trade1.Price) // 先与高价买单成交
	assert.Equal(t, "0.3", trade1.Amount)

	// 验证第二笔成交
	trade2 := result.Trades[1]
	assert.Equal(t, "49500", trade2.Price) // 再与低价买单成交
	assert.Equal(t, "0.7", trade2.Amount)

	// 验证市价单状态
	assert.Equal(t, int64(3), marketSellOrder.Status) // 完全成交
	assert.Equal(t, "1", marketSellOrder.FilledAmount)
}

func TestMatchingEngine_CancelOrder(t *testing.T) {
	engine := NewMatchingEngine()

	order := &model.Order{
		ID:     1,
		UserID: 1,
		Symbol: "BTC/USDT",
		Type:   1,
		Side:   1, // 买单
		Amount: "1.0",
		Price:  "50000.0",
		Status: 1,
	}

	// 添加订单到订单簿
	orderBook := engine.GetOrderBook("BTC/USDT")
	orderBook.AddOrder(order)

	// 验证订单存在
	bestBid, exists := orderBook.GetBestBid()
	assert.True(t, exists)
	assert.Equal(t, "50000", bestBid.Price.String())

	// 取消订单
	err := engine.CancelOrder(order)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), order.Status) // 已取消

	// 验证订单从订单簿中移除
	_, exists = orderBook.GetBestBid()
	assert.False(t, exists)
}

func TestMatchingEngine_GetMarketDepth(t *testing.T) {
	engine := NewMatchingEngine()

	// 添加多个订单
	orders := []*model.Order{
		{ID: 1, Side: 1, Price: "50000.0", Amount: "1.0"},   // 买单
		{ID: 2, Side: 1, Price: "49000.0", Amount: "2.0"},   // 买单
		{ID: 3, Side: 2, Price: "51000.0", Amount: "0.8"},   // 卖单
		{ID: 4, Side: 2, Price: "52000.0", Amount: "1.5"},   // 卖单
	}

	orderBook := engine.GetOrderBook("BTC/USDT")
	for _, order := range orders {
		orderBook.AddOrder(order)
	}

	// 获取市场深度
	bids, asks := engine.GetMarketDepth("BTC/USDT", 2)

	// 验证买盘深度
	assert.Equal(t, 2, len(bids))
	assert.Equal(t, "50000", bids[0].Price.String()) // 最高买价
	assert.Equal(t, "49000", bids[1].Price.String())

	// 验证卖盘深度
	assert.Equal(t, 2, len(asks))
	assert.Equal(t, "51000", asks[0].Price.String()) // 最低卖价
	assert.Equal(t, "52000", asks[1].Price.String())
}

func TestMatchingEngine_GetOrderBookSnapshot(t *testing.T) {
	engine := NewMatchingEngine()

	// 添加一些订单
	buyOrder := &model.Order{ID: 1, Side: 1, Price: "50000.0", Amount: "1.0"}
	sellOrder := &model.Order{ID: 2, Side: 2, Price: "51000.0", Amount: "0.5"}

	orderBook := engine.GetOrderBook("BTC/USDT")
	orderBook.AddOrder(buyOrder)
	orderBook.AddOrder(sellOrder)

	// 获取订单簿快照
	snapshot := engine.GetOrderBookSnapshot("BTC/USDT")

	// 验证快照信息
	assert.Equal(t, "BTC/USDT", snapshot["symbol"])
	assert.Equal(t, 1, snapshot["bid_levels"])
	assert.Equal(t, 1, snapshot["ask_levels"])
	assert.Equal(t, "50000", snapshot["best_bid"])
	assert.Equal(t, "1", snapshot["best_bid_amount"])
	assert.Equal(t, "51000", snapshot["best_ask"])
	assert.Equal(t, "0.5", snapshot["best_ask_amount"])
} 