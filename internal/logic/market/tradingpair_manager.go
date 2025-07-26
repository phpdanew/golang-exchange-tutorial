package market

import (
	"context"
	"fmt"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/model"

	"github.com/zeromicro/go-zero/core/logx"
)

// TradingPairManager 交易对管理器
type TradingPairManager struct {
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	validator *TradingPairValidator
	logx.Logger
}

// NewTradingPairManager 创建交易对管理器
func NewTradingPairManager(ctx context.Context, svcCtx *svc.ServiceContext) *TradingPairManager {
	return &TradingPairManager{
		ctx:       ctx,
		svcCtx:    svcCtx,
		validator: NewTradingPairValidator(),
		Logger:    logx.WithContext(ctx),
	}
}

// InitializeDefaultTradingPairs 初始化默认交易对数据
func (m *TradingPairManager) InitializeDefaultTradingPairs() error {
	// 定义默认交易对配置
	defaultPairs := []model.TradingPair{
		{
			Symbol:        "BTC/USDT",
			BaseCurrency:  "BTC",
			QuoteCurrency: "USDT",
			MinAmount:     "0.00001",
			MaxAmount:     "1000",
			PriceScale:    2,
			AmountScale:   8,
			Status:        1,
			CreatedAt:     time.Now(),
		},
		{
			Symbol:        "ETH/USDT",
			BaseCurrency:  "ETH",
			QuoteCurrency: "USDT",
			MinAmount:     "0.001",
			MaxAmount:     "10000",
			PriceScale:    2,
			AmountScale:   6,
			Status:        1,
			CreatedAt:     time.Now(),
		},
		{
			Symbol:        "BNB/USDT",
			BaseCurrency:  "BNB",
			QuoteCurrency: "USDT",
			MinAmount:     "0.01",
			MaxAmount:     "100000",
			PriceScale:    2,
			AmountScale:   4,
			Status:        1,
			CreatedAt:     time.Now(),
		},
		{
			Symbol:        "ADA/USDT",
			BaseCurrency:  "ADA",
			QuoteCurrency: "USDT",
			MinAmount:     "1",
			MaxAmount:     "1000000",
			PriceScale:    4,
			AmountScale:   2,
			Status:        1,
			CreatedAt:     time.Now(),
		},
		{
			Symbol:        "DOT/USDT",
			BaseCurrency:  "DOT",
			QuoteCurrency: "USDT",
			MinAmount:     "0.1",
			MaxAmount:     "50000",
			PriceScale:    3,
			AmountScale:   3,
			Status:        1,
			CreatedAt:     time.Now(),
		},
	}

	// 插入交易对数据（如果不存在）
	for _, pair := range defaultPairs {
		// 检查交易对是否已存在
		existing, err := m.svcCtx.TradingPairModel.FindBySymbol(m.ctx, pair.Symbol)
		if err != nil && err != model.ErrNotFound {
			m.Logger.Errorf("Failed to check existing trading pair %s: %v", pair.Symbol, err)
			return err
		}

		// 如果不存在则插入
		if existing == nil {
			_, err := m.svcCtx.TradingPairModel.Insert(m.ctx, &pair)
			if err != nil {
				m.Logger.Errorf("Failed to insert trading pair %s: %v", pair.Symbol, err)
				return err
			}
			m.Logger.Infof("Initialized trading pair: %s", pair.Symbol)
		} else {
			m.Logger.Infof("Trading pair already exists: %s", pair.Symbol)
		}
	}

	return nil
}

// CreateTradingPair 创建新的交易对
func (m *TradingPairManager) CreateTradingPair(pair *model.TradingPair) error {
	// 验证交易对数据
	if err := m.validateTradingPair(pair); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 检查交易对是否已存在
	existing, err := m.svcCtx.TradingPairModel.FindBySymbol(m.ctx, pair.Symbol)
	if err != nil && err != model.ErrNotFound {
		return fmt.Errorf("failed to check existing trading pair: %w", err)
	}

	if existing != nil {
		return fmt.Errorf("trading pair %s already exists", pair.Symbol)
	}

	// 设置创建时间
	pair.CreatedAt = time.Now()

	// 插入数据库
	_, err = m.svcCtx.TradingPairModel.Insert(m.ctx, pair)
	if err != nil {
		return fmt.Errorf("failed to create trading pair: %w", err)
	}

	m.Logger.Infof("Created trading pair: %s", pair.Symbol)
	return nil
}

// UpdateTradingPairStatus 更新交易对状态
func (m *TradingPairManager) UpdateTradingPairStatus(symbol string, status int64) error {
	// 验证状态
	if err := m.validator.ValidateStatus(status); err != nil {
		return err
	}

	// 获取现有交易对
	pair, err := m.svcCtx.TradingPairModel.FindBySymbol(m.ctx, symbol)
	if err != nil {
		if err == model.ErrNotFound {
			return fmt.Errorf("trading pair not found: %s", symbol)
		}
		return fmt.Errorf("failed to get trading pair: %w", err)
	}

	// 更新状态
	pair.Status = status

	// 保存到数据库
	err = m.svcCtx.TradingPairModel.Update(m.ctx, pair)
	if err != nil {
		return fmt.Errorf("failed to update trading pair status: %w", err)
	}

	statusText := "disabled"
	if status == 1 {
		statusText = "active"
	}
	m.Logger.Infof("Updated trading pair %s status to %s", symbol, statusText)
	return nil
}

// GetActiveTradingPairs 获取所有活跃的交易对
func (m *TradingPairManager) GetActiveTradingPairs() ([]*model.TradingPair, error) {
	pairs, err := m.svcCtx.TradingPairModel.FindActivePairs(m.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active trading pairs: %w", err)
	}
	return pairs, nil
}

// GetTradingPairBySymbol 根据符号获取交易对
func (m *TradingPairManager) GetTradingPairBySymbol(symbol string) (*model.TradingPair, error) {
	if err := m.validator.ValidateSymbol(symbol); err != nil {
		return nil, err
	}

	pair, err := m.svcCtx.TradingPairModel.FindBySymbol(m.ctx, symbol)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("trading pair not found: %s", symbol)
		}
		return nil, fmt.Errorf("failed to get trading pair: %w", err)
	}

	return pair, nil
}

// ValidateOrderForTradingPair 验证订单是否符合交易对要求
func (m *TradingPairManager) ValidateOrderForTradingPair(symbol, amount, price string, orderType int64) error {
	// 获取交易对信息
	pair, err := m.GetTradingPairBySymbol(symbol)
	if err != nil {
		return err
	}

	// 检查交易对是否可用
	if pair.Status != 1 {
		return fmt.Errorf("trading pair %s is currently disabled", symbol)
	}

	// 验证订单数量
	if err := m.validator.ValidateOrderAmount(amount, pair.MinAmount, pair.MaxAmount, pair.AmountScale); err != nil {
		return err
	}

	// 如果是限价单，验证价格精度
	if orderType == 1 && price != "" {
		if err := m.validator.ValidateOrderPrice(price, pair.PriceScale); err != nil {
			return err
		}
	}

	return nil
}

// validateTradingPair 验证交易对数据的完整性
func (m *TradingPairManager) validateTradingPair(pair *model.TradingPair) error {
	// 验证交易对符号
	if err := m.validator.ValidateSymbol(pair.Symbol); err != nil {
		return err
	}

	// 验证基础币种和计价币种
	if err := m.validator.ValidateCurrency(pair.BaseCurrency); err != nil {
		return fmt.Errorf("invalid base currency: %w", err)
	}

	if err := m.validator.ValidateCurrency(pair.QuoteCurrency); err != nil {
		return fmt.Errorf("invalid quote currency: %w", err)
	}

	// 验证最小和最大交易数量
	if err := m.validator.ValidateMinMaxAmount(pair.MinAmount, pair.MaxAmount); err != nil {
		return err
	}

	// 验证精度设置
	if err := m.validator.ValidateScale(pair.PriceScale, "price_scale"); err != nil {
		return err
	}

	if err := m.validator.ValidateScale(pair.AmountScale, "amount_scale"); err != nil {
		return err
	}

	// 验证状态
	if err := m.validator.ValidateStatus(pair.Status); err != nil {
		return err
	}

	return nil
}

// GetTradingPairStats 获取交易对统计信息
func (m *TradingPairManager) GetTradingPairStats() (map[string]interface{}, error) {
	// 获取所有交易对
	allPairs, err := m.svcCtx.TradingPairModel.FindByStatus(m.ctx, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get trading pairs: %w", err)
	}

	activePairs, err := m.svcCtx.TradingPairModel.FindActivePairs(m.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active trading pairs: %w", err)
	}

	disabledPairs, err := m.svcCtx.TradingPairModel.FindByStatus(m.ctx, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to get disabled trading pairs: %w", err)
	}

	stats := map[string]interface{}{
		"total_pairs":    len(allPairs),
		"active_pairs":   len(activePairs),
		"disabled_pairs": len(disabledPairs),
		"supported_base_currencies":  m.getUniqueCurrencies(allPairs, "base"),
		"supported_quote_currencies": m.getUniqueCurrencies(allPairs, "quote"),
	}

	return stats, nil
}

// getUniqueCurrencies 获取唯一的币种列表
func (m *TradingPairManager) getUniqueCurrencies(pairs []*model.TradingPair, currencyType string) []string {
	currencyMap := make(map[string]bool)
	var currencies []string

	for _, pair := range pairs {
		var currency string
		if currencyType == "base" {
			currency = pair.BaseCurrency
		} else {
			currency = pair.QuoteCurrency
		}

		if !currencyMap[currency] {
			currencyMap[currency] = true
			currencies = append(currencies, currency)
		}
	}

	return currencies
}