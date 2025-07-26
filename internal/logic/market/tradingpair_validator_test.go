package market

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTradingPairValidator_ValidateSymbol(t *testing.T) {
	validator := NewTradingPairValidator()

	tests := []struct {
		name    string
		symbol  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid symbol BTC/USDT",
			symbol:  "BTC/USDT",
			wantErr: false,
		},
		{
			name:    "valid symbol ETH/BTC",
			symbol:  "ETH/BTC",
			wantErr: false,
		},
		{
			name:    "empty symbol",
			symbol:  "",
			wantErr: true,
			errMsg:  "trading pair symbol cannot be empty",
		},
		{
			name:    "invalid format - no slash",
			symbol:  "BTCUSDT",
			wantErr: true,
			errMsg:  "invalid trading pair symbol format",
		},
		{
			name:    "invalid format - lowercase",
			symbol:  "btc/usdt",
			wantErr: true,
			errMsg:  "invalid trading pair symbol format",
		},
		{
			name:    "invalid format - same currencies",
			symbol:  "BTC/BTC",
			wantErr: true,
			errMsg:  "base currency and quote currency cannot be the same",
		},
		{
			name:    "invalid format - too short currency",
			symbol:  "B/USDT",
			wantErr: true,
			errMsg:  "invalid trading pair symbol format",
		},
		{
			name:    "invalid format - too long currency",
			symbol:  "VERYLONGCURRENCY/USDT",
			wantErr: true,
			errMsg:  "invalid trading pair symbol format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSymbol(tt.symbol)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTradingPairValidator_ValidateCurrency(t *testing.T) {
	validator := NewTradingPairValidator()

	tests := []struct {
		name     string
		currency string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid currency BTC",
			currency: "BTC",
			wantErr:  false,
		},
		{
			name:     "valid currency USDT",
			currency: "USDT",
			wantErr:  false,
		},
		{
			name:     "empty currency",
			currency: "",
			wantErr:  true,
			errMsg:   "currency cannot be empty",
		},
		{
			name:     "lowercase currency",
			currency: "btc",
			wantErr:  true,
			errMsg:   "invalid currency format",
		},
		{
			name:     "too short currency",
			currency: "B",
			wantErr:  true,
			errMsg:   "invalid currency format",
		},
		{
			name:     "too long currency",
			currency: "VERYLONGCURRENCY",
			wantErr:  true,
			errMsg:   "invalid currency format",
		},
		{
			name:     "currency with numbers",
			currency: "BTC1",
			wantErr:  true,
			errMsg:   "invalid currency format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCurrency(tt.currency)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTradingPairValidator_ValidateAmount(t *testing.T) {
	validator := NewTradingPairValidator()

	tests := []struct {
		name      string
		amount    string
		fieldName string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid amount",
			amount:    "1.5",
			fieldName: "amount",
			wantErr:   false,
		},
		{
			name:      "valid small amount",
			amount:    "0.00000001",
			fieldName: "amount",
			wantErr:   false,
		},
		{
			name:      "empty amount",
			amount:    "",
			fieldName: "amount",
			wantErr:   true,
			errMsg:    "amount cannot be empty",
		},
		{
			name:      "invalid amount format",
			amount:    "abc",
			fieldName: "amount",
			wantErr:   true,
			errMsg:    "invalid amount format",
		},
		{
			name:      "zero amount",
			amount:    "0",
			fieldName: "amount",
			wantErr:   true,
			errMsg:    "amount must be positive",
		},
		{
			name:      "negative amount",
			amount:    "-1.5",
			fieldName: "amount",
			wantErr:   true,
			errMsg:    "amount must be positive",
		},
		{
			name:      "too high precision",
			amount:    "1.1234567890123456789",
			fieldName: "amount",
			wantErr:   true,
			errMsg:    "precision cannot exceed 18 decimal places",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAmount(tt.amount, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTradingPairValidator_ValidateMinMaxAmount(t *testing.T) {
	validator := NewTradingPairValidator()

	tests := []struct {
		name      string
		minAmount string
		maxAmount string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid min max amounts",
			minAmount: "0.001",
			maxAmount: "1000",
			wantErr:   false,
		},
		{
			name:      "invalid min amount",
			minAmount: "abc",
			maxAmount: "1000",
			wantErr:   true,
			errMsg:    "invalid min_amount format",
		},
		{
			name:      "invalid max amount",
			minAmount: "0.001",
			maxAmount: "xyz",
			wantErr:   true,
			errMsg:    "invalid max_amount format",
		},
		{
			name:      "max amount less than min amount",
			minAmount: "1000",
			maxAmount: "0.001",
			wantErr:   true,
			errMsg:    "max_amount must be greater than min_amount",
		},
		{
			name:      "max amount equal to min amount",
			minAmount: "1.0",
			maxAmount: "1.0",
			wantErr:   true,
			errMsg:    "max_amount must be greater than min_amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMinMaxAmount(tt.minAmount, tt.maxAmount)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTradingPairValidator_ValidateScale(t *testing.T) {
	validator := NewTradingPairValidator()

	tests := []struct {
		name      string
		scale     int64
		fieldName string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid scale 8",
			scale:     8,
			fieldName: "price_scale",
			wantErr:   false,
		},
		{
			name:      "valid scale 0",
			scale:     0,
			fieldName: "amount_scale",
			wantErr:   false,
		},
		{
			name:      "valid scale 18",
			scale:     18,
			fieldName: "price_scale",
			wantErr:   false,
		},
		{
			name:      "negative scale",
			scale:     -1,
			fieldName: "price_scale",
			wantErr:   true,
			errMsg:    "price_scale must be between 0 and 18",
		},
		{
			name:      "scale too high",
			scale:     19,
			fieldName: "amount_scale",
			wantErr:   true,
			errMsg:    "amount_scale must be between 0 and 18",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateScale(tt.scale, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTradingPairValidator_ValidateStatus(t *testing.T) {
	validator := NewTradingPairValidator()

	tests := []struct {
		name    string
		status  int64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid status active",
			status:  1,
			wantErr: false,
		},
		{
			name:    "valid status disabled",
			status:  2,
			wantErr: false,
		},
		{
			name:    "invalid status 0",
			status:  0,
			wantErr: true,
			errMsg:  "invalid status, must be 1 (active) or 2 (disabled)",
		},
		{
			name:    "invalid status 3",
			status:  3,
			wantErr: true,
			errMsg:  "invalid status, must be 1 (active) or 2 (disabled)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStatus(tt.status)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTradingPairValidator_ValidateOrderAmount(t *testing.T) {
	validator := NewTradingPairValidator()

	tests := []struct {
		name        string
		amount      string
		minAmount   string
		maxAmount   string
		amountScale int64
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid order amount",
			amount:      "1.5",
			minAmount:   "0.001",
			maxAmount:   "1000",
			amountScale: 8,
			wantErr:     false,
		},
		{
			name:        "amount below minimum",
			amount:      "0.0001",
			minAmount:   "0.001",
			maxAmount:   "1000",
			amountScale: 8,
			wantErr:     true,
			errMsg:      "is less than minimum",
		},
		{
			name:        "amount above maximum",
			amount:      "2000",
			minAmount:   "0.001",
			maxAmount:   "1000",
			amountScale: 8,
			wantErr:     true,
			errMsg:      "exceeds maximum",
		},
		{
			name:        "amount precision too high",
			amount:      "1.123456789",
			minAmount:   "0.001",
			maxAmount:   "1000",
			amountScale: 6,
			wantErr:     true,
			errMsg:      "precision exceeds 6 decimal places",
		},
		{
			name:        "invalid amount format",
			amount:      "abc",
			minAmount:   "0.001",
			maxAmount:   "1000",
			amountScale: 8,
			wantErr:     true,
			errMsg:      "invalid order amount format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOrderAmount(tt.amount, tt.minAmount, tt.maxAmount, tt.amountScale)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTradingPairValidator_ValidateOrderPrice(t *testing.T) {
	validator := NewTradingPairValidator()

	tests := []struct {
		name       string
		price      string
		priceScale int64
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid price",
			price:      "50000.12",
			priceScale: 2,
			wantErr:    false,
		},
		{
			name:       "empty price",
			price:      "",
			priceScale: 2,
			wantErr:    true,
			errMsg:     "price cannot be empty",
		},
		{
			name:       "invalid price format",
			price:      "abc",
			priceScale: 2,
			wantErr:    true,
			errMsg:     "invalid price format",
		},
		{
			name:       "zero price",
			price:      "0",
			priceScale: 2,
			wantErr:    true,
			errMsg:     "price must be positive",
		},
		{
			name:       "negative price",
			price:      "-100",
			priceScale: 2,
			wantErr:    true,
			errMsg:     "price must be positive",
		},
		{
			name:       "price precision too high",
			price:      "50000.123",
			priceScale: 2,
			wantErr:    true,
			errMsg:     "precision exceeds 2 decimal places",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOrderPrice(tt.price, tt.priceScale)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}