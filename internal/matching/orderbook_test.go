package matching

import (
	"testing"
	"time"

	"crypto-exchange/model"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestOrderBook_AddOrder(t *testing.T) {
	orderBook := NewOrderBook("BTC/USDT")

	// 创建测试订单
	buyOrder := &model.Order{
		ID:     1,
		UserID: 1,
		Symbol: "BTC/USDT",
		Type:   1, // 限价单
		Side:   1, // 买单
		Amount: "1.0",
		Price:  "50000.0",
		Status: 1,
	}

	sellOrder := &model.Order{
		ID:     2,
		UserID: 2,
		Symbol: "BTC/USDT",
		Type:   1, // 限价单
		Side:   2, // 卖单
		Amount: "0.5",
		Price:  "51000.0",
		Status: 1,
	}

	// 添加订单
	orderBook.AddOrder(buyOrder)
	orderBook.AddOrder(sellOrder)

	// 验证买盘
	assert.Equal(t, 1, len(orderBook.Bids))
	assert.Equal(t, 1, len(orderBook.BidPrices))
	bestBid, exists := orderBook.GetBestBid()
	assert.True(t, exists)
	assert.Equal(t, "50000", bestBid.Price.String())
	assert.Equal(t, "1", bestBid.Total.String())

	// 验证卖盘
	assert.Equal(t, 1, len(orderBook.Asks))
	assert.Equal(t, 1, len(orderBook.AskPrices))
	bestAsk, exists := orderBook.GetBestAsk()
	assert.True(t, exists)
	assert.Equal(t, "51000", bestAsk.Price.String())
	assert.Equal(t, "0.5", bestAsk.Total.String())
}

func TestOrderBook_PricePriority(t *testing.T) {
	orderBook := NewOrderBook("BTC/USDT")

	// 添加多个不同价格的买单
	orders := []*model.Order{
		{ID: 1, Side: 1, Price: "50000.0", Amount: "1.0"},
		{ID: 2, Side: 1, Price: "51000.0", Amount: "0.5"}, // 更高价格
		{ID: 3, Side: 1, Price: "49000.0", Amount: "2.0"}, // 更低价格
	}

	for _, order := range orders {
		orderBook.AddOrder(order)
	}

	// 验证买盘价格排序（从高到低）
	assert.Equal(t, 3, len(orderBook.BidPrices))
	assert.Equal(t, "51000", orderBook.BidPrices[0].String()) // 最高价在第一位
	assert.Equal(t, "50000", orderBook.BidPrices[1].String())
	assert.Equal(t, "49000", orderBook.BidPrices[2].String()) // 最低价在最后

	// 验证最优买价
	bestBid, exists := orderBook.GetBestBid()
	assert.True(t, exists)
	assert.Equal(t, "51000", bestBid.Price.String())

	// 添加多个不同价格的卖单
	sellOrders := []*model.Order{
		{ID: 4, Side: 2, Price: "52000.0", Amount: "1.0"},
		{ID: 5, Side: 2, Price: "51500.0", Amount: "0.5"}, // 更低价格
		{ID: 6, Side: 2, Price: "53000.0", Amount: "2.0"}, // 更高价格
	}

	for _, order := range sellOrders {
		orderBook.AddOrder(order)
	}

	// 验证卖盘价格排序（从低到高）
	assert.Equal(t, 3, len(orderBook.AskPrices))
	assert.Equal(t, "51500", orderBook.AskPrices[0].String()) // 最低价在第一位
	assert.Equal(t, "52000", orderBook.AskPrices[1].String())
	assert.Equal(t, "53000", orderBook.AskPrices[2].String()) // 最高价在最后

	// 验证最优卖价
	bestAsk, exists := orderBook.GetBestAsk()
	assert.True(t, exists)
	assert.Equal(t, "51500", bestAsk.Price.String())
}

func TestOrderBook_TimePriority(t *testing.T) {
	orderBook := NewOrderBook("BTC/USDT")

	// 添加相同价格的多个订单（时间优先）
	order1 := &model.Order{ID: 1, Side: 1, Price: "50000.0", Amount: "1.0", CreatedAt: time.Now()}
	time.Sleep(time.Millisecond) // 确保时间不同
	order2 := &model.Order{ID: 2, Side: 1, Price: "50000.0", Amount: "0.5", CreatedAt: time.Now()}

	orderBook.AddOrder(order1)
	orderBook.AddOrder(order2)

	// 验证同一价格层级中订单的排序（先进先出）
	priceLevel := orderBook.Bids["50000"]
	assert.Equal(t, 2, priceLevel.Orders.Len())
	assert.Equal(t, "1.5", priceLevel.Total.String()) // 总数量

	// 第一个订单应该在队列前面
	firstOrder := priceLevel.Orders.Front().Value.(*model.Order)
	assert.Equal(t, uint64(1), firstOrder.ID)
}

func TestOrderBook_RemoveOrder(t *testing.T) {
	orderBook := NewOrderBook("BTC/USDT")

	order := &model.Order{
		ID:     1,
		Side:   1, // 买单
		Price:  "50000.0",
		Amount: "1.0",
	}

	// 添加订单
	orderBook.AddOrder(order)
	assert.Equal(t, 1, len(orderBook.Bids))

	// 移除订单
	orderBook.RemoveOrder(order)
	assert.Equal(t, 0, len(orderBook.Bids))
	assert.Equal(t, 0, len(orderBook.BidPrices))

	// 验证无法获取最优买价
	_, exists := orderBook.GetBestBid()
	assert.False(t, exists)
}

func TestOrderBook_UpdateOrderAmount(t *testing.T) {
	orderBook := NewOrderBook("BTC/USDT")

	order := &model.Order{
		ID:     1,
		Side:   1, // 买单
		Price:  "50000.0",
		Amount: "1.0",
	}

	orderBook.AddOrder(order)

	// 验证初始数量
	priceLevel := orderBook.Bids["50000"]
	assert.Equal(t, "1", priceLevel.Total.String())

	// 更新订单数量
	newAmount := decimal.NewFromFloat(0.5)
	orderBook.UpdateOrderAmount(order, newAmount)

	// 验证更新后的数量
	assert.Equal(t, "0.5", priceLevel.Total.String())
	assert.Equal(t, "0.5", order.Amount)
}

func TestOrderBook_GetDepth(t *testing.T) {
	orderBook := NewOrderBook("BTC/USDT")

	// 添加多个买单和卖单
	buyOrders := []*model.Order{
		{ID: 1, Side: 1, Price: "50000.0", Amount: "1.0"},
		{ID: 2, Side: 1, Price: "49000.0", Amount: "2.0"},
		{ID: 3, Side: 1, Price: "48000.0", Amount: "1.5"},
	}

	sellOrders := []*model.Order{
		{ID: 4, Side: 2, Price: "51000.0", Amount: "0.8"},
		{ID: 5, Side: 2, Price: "52000.0", Amount: "1.2"},
		{ID: 6, Side: 2, Price: "53000.0", Amount: "0.5"},
	}

	for _, order := range buyOrders {
		orderBook.AddOrder(order)
	}
	for _, order := range sellOrders {
		orderBook.AddOrder(order)
	}

	// 获取深度为2的市场深度
	bids, asks := orderBook.GetDepth(2)

	// 验证买盘深度（前2个最高价）
	assert.Equal(t, 2, len(bids))
	assert.Equal(t, "50000", bids[0].Price.String())
	assert.Equal(t, "1", bids[0].Total.String())
	assert.Equal(t, "49000", bids[1].Price.String())
	assert.Equal(t, "2", bids[1].Total.String())

	// 验证卖盘深度（前2个最低价）
	assert.Equal(t, 2, len(asks))
	assert.Equal(t, "51000", asks[0].Price.String())
	assert.Equal(t, "0.8", asks[0].Total.String())
	assert.Equal(t, "52000", asks[1].Price.String())
	assert.Equal(t, "1.2", asks[1].Total.String())
}

func TestOrderBook_Clear(t *testing.T) {
	orderBook := NewOrderBook("BTC/USDT")

	// 添加一些订单
	order1 := &model.Order{ID: 1, Side: 1, Price: "50000.0", Amount: "1.0"}
	order2 := &model.Order{ID: 2, Side: 2, Price: "51000.0", Amount: "0.5"}

	orderBook.AddOrder(order1)
	orderBook.AddOrder(order2)

	assert.Equal(t, 1, len(orderBook.Bids))
	assert.Equal(t, 1, len(orderBook.Asks))

	// 清空订单簿
	orderBook.Clear()

	// 验证清空后的状态
	assert.Equal(t, 0, len(orderBook.Bids))
	assert.Equal(t, 0, len(orderBook.Asks))
	assert.Equal(t, 0, len(orderBook.BidPrices))
	assert.Equal(t, 0, len(orderBook.AskPrices))

	// 验证无法获取最优价格
	_, exists := orderBook.GetBestBid()
	assert.False(t, exists)
	_, exists = orderBook.GetBestAsk()
	assert.False(t, exists)
} 