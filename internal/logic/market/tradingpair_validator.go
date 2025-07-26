package market

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
)

// TradingPairValidator 交易对验证器
type TradingPairValidator struct{}

// NewTradingPairValidator 创建交易对验证器
func NewTradingPairValidator() *TradingPairValidator {
	return &TradingPairValidator{}
}

// ValidateSymbol 验证交易对符号格式
func (v *TradingPairValidator) ValidateSymbol(symbol string) error {
	if symbol == "" {
		return fmt.Errorf("trading pair symbol cannot be empty")
	}

	// 验证符号格式：基础币种/计价币种，如BTC/USDT
	symbolPattern := regexp.MustCompile(`^[A-Z]{2,10}/[A-Z]{2,10}$`)
	if !symbolPattern.MatchString(symbol) {
		return fmt.Errorf("invalid trading pair symbol format, expected format: BASE/QUOTE (e.g., BTC/USDT)")
	}

	// 检查基础币种和计价币种不能相同
	parts := strings.Split(symbol, "/")
	if parts[0] == parts[1] {
		return fmt.Errorf("base currency and quote currency cannot be the same")
	}

	return nil
}

// ValidateCurrency 验证币种代码
func (v *TradingPairValidator) ValidateCurrency(currency string) error {
	if currency == "" {
		return fmt.Errorf("currency cannot be empty")
	}

	// 币种代码应该是2-10位大写字母
	currencyPattern := regexp.MustCompile(`^[A-Z]{2,10}$`)
	if !currencyPattern.MatchString(currency) {
		return fmt.Errorf("invalid currency format, should be 2-10 uppercase letters")
	}

	return nil
}

// ValidateAmount 验证交易数量
func (v *TradingPairValidator) ValidateAmount(amount string, fieldName string) error {
	if amount == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	// 转换为decimal进行验证
	amountDecimal, err := decimal.NewFromString(amount)
	if err != nil {
		return fmt.Errorf("invalid %s format: %s", fieldName, amount)
	}

	// 数量必须为正数
	if amountDecimal.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("%s must be positive", fieldName)
	}

	// 检查精度不超过18位小数
	if amountDecimal.Exponent() < -18 {
		return fmt.Errorf("%s precision cannot exceed 18 decimal places", fieldName)
	}

	return nil
}

// ValidateMinMaxAmount 验证最小和最大交易数量的关系
func (v *TradingPairValidator) ValidateMinMaxAmount(minAmount, maxAmount string) error {
	if err := v.ValidateAmount(minAmount, "min_amount"); err != nil {
		return err
	}

	if err := v.ValidateAmount(maxAmount, "max_amount"); err != nil {
		return err
	}

	minDecimal, _ := decimal.NewFromString(minAmount)
	maxDecimal, _ := decimal.NewFromString(maxAmount)

	// 最大数量必须大于最小数量
	if maxDecimal.LessThanOrEqual(minDecimal) {
		return fmt.Errorf("max_amount must be greater than min_amount")
	}

	return nil
}

// ValidateScale 验证精度设置
func (v *TradingPairValidator) ValidateScale(scale int64, fieldName string) error {
	// 精度必须在0-18之间
	if scale < 0 || scale > 18 {
		return fmt.Errorf("%s must be between 0 and 18", fieldName)
	}

	return nil
}

// ValidateStatus 验证交易对状态
func (v *TradingPairValidator) ValidateStatus(status int64) error {
	// 状态：1-正常交易，2-禁用交易
	if status != 1 && status != 2 {
		return fmt.Errorf("invalid status, must be 1 (active) or 2 (disabled)")
	}

	return nil
}

// ValidateOrderAmount 验证订单数量是否符合交易对限制
func (v *TradingPairValidator) ValidateOrderAmount(amount, minAmount, maxAmount string, amountScale int64) error {
	if err := v.ValidateAmount(amount, "order amount"); err != nil {
		return err
	}

	amountDecimal, _ := decimal.NewFromString(amount)
	minDecimal, _ := decimal.NewFromString(minAmount)
	maxDecimal, _ := decimal.NewFromString(maxAmount)

	// 检查数量是否在允许范围内
	if amountDecimal.LessThan(minDecimal) {
		return fmt.Errorf("order amount %s is less than minimum %s", amount, minAmount)
	}

	if amountDecimal.GreaterThan(maxDecimal) {
		return fmt.Errorf("order amount %s exceeds maximum %s", amount, maxAmount)
	}

	// 检查数量精度
	if amountDecimal.Exponent() < -int32(amountScale) {
		return fmt.Errorf("order amount precision exceeds %d decimal places", amountScale)
	}

	return nil
}

// ValidateOrderPrice 验证订单价格精度
func (v *TradingPairValidator) ValidateOrderPrice(price string, priceScale int64) error {
	if price == "" {
		return fmt.Errorf("price cannot be empty")
	}

	priceDecimal, err := decimal.NewFromString(price)
	if err != nil {
		return fmt.Errorf("invalid price format: %s", price)
	}

	// 价格必须为正数
	if priceDecimal.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("price must be positive")
	}

	// 检查价格精度
	if priceDecimal.Exponent() < -int32(priceScale) {
		return fmt.Errorf("price precision exceeds %d decimal places", priceScale)
	}

	return nil
}