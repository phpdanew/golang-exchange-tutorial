package matching

import (
	"container/list"
	"sort"
	"sync"
	"time"

	"crypto-exchange/model"

	"github.com/shopspring/decimal"
)

// PriceLevel 价格层级，包含某个价格的所有订单
type PriceLevel struct {
	Price  decimal.Decimal   // 价格
	Orders *list.List        // 该价格的订单队列，按时间优先排序
	Total  decimal.Decimal   // 该价格层级的总数量
}

// NewPriceLevel 创建新的价格层级
func NewPriceLevel(price decimal.Decimal) *PriceLevel {
	return &PriceLevel{
		Price:  price,
		Orders: list.New(),
		Total:  decimal.Zero,
	}
}

// AddOrder 添加订单到价格层级
func (pl *PriceLevel) AddOrder(order *model.Order) {
	pl.Orders.PushBack(order)
	amount, _ := decimal.NewFromString(order.Amount)
	pl.Total = pl.Total.Add(amount)
}

// RemoveOrder 从价格层级移除订单
func (pl *PriceLevel) RemoveOrder(order *model.Order) {
	for e := pl.Orders.Front(); e != nil; e = e.Next() {
		if e.Value.(*model.Order).ID == order.ID {
			pl.Orders.Remove(e)
			amount, _ := decimal.NewFromString(order.Amount)
			pl.Total = pl.Total.Sub(amount)
			break
		}
	}
}

// UpdateOrderAmount 更新订单数量
func (pl *PriceLevel) UpdateOrderAmount(order *model.Order, newAmount decimal.Decimal) {
	oldAmount, _ := decimal.NewFromString(order.Amount)
	pl.Total = pl.Total.Sub(oldAmount).Add(newAmount)
	order.Amount = newAmount.String()
}

// IsEmpty 检查价格层级是否为空
func (pl *PriceLevel) IsEmpty() bool {
	return pl.Orders.Len() == 0
}

// OrderBook 订单簿数据结构
type OrderBook struct {
	Symbol     string                       // 交易对符号
	Bids       map[string]*PriceLevel      // 买盘，价格->价格层级
	Asks       map[string]*PriceLevel      // 卖盘，价格->价格层级
	BidPrices  []decimal.Decimal           // 买盘价格列表，按价格从高到低排序
	AskPrices  []decimal.Decimal           // 卖盘价格列表，按价格从低到高排序
	mutex      sync.RWMutex                // 读写锁，保证并发安全
	LastUpdate time.Time                   // 最后更新时间
}

// NewOrderBook 创建新的订单簿
func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol:     symbol,
		Bids:       make(map[string]*PriceLevel),
		Asks:       make(map[string]*PriceLevel),
		BidPrices:  make([]decimal.Decimal, 0),
		AskPrices:  make([]decimal.Decimal, 0),
		LastUpdate: time.Now(),
	}
}

// AddOrder 添加订单到订单簿
func (ob *OrderBook) AddOrder(order *model.Order) {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	price, _ := decimal.NewFromString(order.Price)
	priceStr := price.String()

	if order.Side == 1 { // 买单
		if _, exists := ob.Bids[priceStr]; !exists {
			ob.Bids[priceStr] = NewPriceLevel(price)
			ob.insertBidPrice(price)
		}
		ob.Bids[priceStr].AddOrder(order)
	} else { // 卖单
		if _, exists := ob.Asks[priceStr]; !exists {
			ob.Asks[priceStr] = NewPriceLevel(price)
			ob.insertAskPrice(price)
		}
		ob.Asks[priceStr].AddOrder(order)
	}

	ob.LastUpdate = time.Now()
}

// RemoveOrder 从订单簿移除订单
func (ob *OrderBook) RemoveOrder(order *model.Order) {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	price, _ := decimal.NewFromString(order.Price)
	priceStr := price.String()

	if order.Side == 1 { // 买单
		if level, exists := ob.Bids[priceStr]; exists {
			level.RemoveOrder(order)
			if level.IsEmpty() {
				delete(ob.Bids, priceStr)
				ob.removeBidPrice(price)
			}
		}
	} else { // 卖单
		if level, exists := ob.Asks[priceStr]; exists {
			level.RemoveOrder(order)
			if level.IsEmpty() {
				delete(ob.Asks, priceStr)
				ob.removeAskPrice(price)
			}
		}
	}

	ob.LastUpdate = time.Now()
}

// GetBestBid 获取最优买价（最高价）
func (ob *OrderBook) GetBestBid() (*PriceLevel, bool) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	if len(ob.BidPrices) == 0 {
		return nil, false
	}

	bestPrice := ob.BidPrices[0]
	level := ob.Bids[bestPrice.String()]
	return level, true
}

// GetBestAsk 获取最优卖价（最低价）
func (ob *OrderBook) GetBestAsk() (*PriceLevel, bool) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	if len(ob.AskPrices) == 0 {
		return nil, false
	}

	bestPrice := ob.AskPrices[0]
	level := ob.Asks[bestPrice.String()]
	return level, true
}

// GetDepth 获取市场深度，返回指定深度的买卖盘数据
func (ob *OrderBook) GetDepth(depth int) ([]PriceLevel, []PriceLevel) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	// 获取买盘深度（价格从高到低）
	bids := make([]PriceLevel, 0, depth)
	for i := 0; i < len(ob.BidPrices) && i < depth; i++ {
		price := ob.BidPrices[i]
		level := ob.Bids[price.String()]
		bids = append(bids, *level)
	}

	// 获取卖盘深度（价格从低到高）
	asks := make([]PriceLevel, 0, depth)
	for i := 0; i < len(ob.AskPrices) && i < depth; i++ {
		price := ob.AskPrices[i]
		level := ob.Asks[price.String()]
		asks = append(asks, *level)
	}

	return bids, asks
}

// insertBidPrice 插入买盘价格，维持从高到低排序
func (ob *OrderBook) insertBidPrice(price decimal.Decimal) {
	ob.BidPrices = append(ob.BidPrices, price)
	sort.Slice(ob.BidPrices, func(i, j int) bool {
		return ob.BidPrices[i].GreaterThan(ob.BidPrices[j]) // 价格从高到低
	})
}

// insertAskPrice 插入卖盘价格，维持从低到高排序
func (ob *OrderBook) insertAskPrice(price decimal.Decimal) {
	ob.AskPrices = append(ob.AskPrices, price)
	sort.Slice(ob.AskPrices, func(i, j int) bool {
		return ob.AskPrices[i].LessThan(ob.AskPrices[j]) // 价格从低到高
	})
}

// removeBidPrice 移除买盘价格
func (ob *OrderBook) removeBidPrice(price decimal.Decimal) {
	for i, p := range ob.BidPrices {
		if p.Equal(price) {
			ob.BidPrices = append(ob.BidPrices[:i], ob.BidPrices[i+1:]...)
			break
		}
	}
}

// removeAskPrice 移除卖盘价格
func (ob *OrderBook) removeAskPrice(price decimal.Decimal) {
	for i, p := range ob.AskPrices {
		if p.Equal(price) {
			ob.AskPrices = append(ob.AskPrices[:i], ob.AskPrices[i+1:]...)
			break
		}
	}
}

// UpdateOrderAmount 更新订单数量
func (ob *OrderBook) UpdateOrderAmount(order *model.Order, newAmount decimal.Decimal) {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	price, _ := decimal.NewFromString(order.Price)
	priceStr := price.String()

	if order.Side == 1 { // 买单
		if level, exists := ob.Bids[priceStr]; exists {
			level.UpdateOrderAmount(order, newAmount)
		}
	} else { // 卖单
		if level, exists := ob.Asks[priceStr]; exists {
			level.UpdateOrderAmount(order, newAmount)
		}
	}

	ob.LastUpdate = time.Now()
}

// Clear 清空订单簿
func (ob *OrderBook) Clear() {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	ob.Bids = make(map[string]*PriceLevel)
	ob.Asks = make(map[string]*PriceLevel)
	ob.BidPrices = make([]decimal.Decimal, 0)
	ob.AskPrices = make([]decimal.Decimal, 0)
	ob.LastUpdate = time.Now()
} 