package model

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestDecimalPrecision(t *testing.T) {
	// Test decimal precision for financial calculations
	tests := []struct {
		name     string
		amount1  string
		amount2  string
		expected string
		op       string
	}{
		{
			name:     "Addition with high precision",
			amount1:  "0.12345678",
			amount2:  "0.87654322",
			expected: "1.00000000",
			op:       "add",
		},
		{
			name:     "Subtraction with high precision",
			amount1:  "1.00000000",
			amount2:  "0.12345678",
			expected: "0.87654322",
			op:       "sub",
		},
		{
			name:     "Multiplication with precision",
			amount1:  "0.1",
			amount2:  "0.2",
			expected: "0.02",
			op:       "mul",
		},
		{
			name:     "Division with precision",
			amount1:  "1",
			amount2:  "3",
			expected: "0.33333333",
			op:       "div",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d1, err := decimal.NewFromString(tt.amount1)
			if err != nil {
				t.Fatalf("Failed to parse amount1: %v", err)
			}

			d2, err := decimal.NewFromString(tt.amount2)
			if err != nil {
				t.Fatalf("Failed to parse amount2: %v", err)
			}

			var result decimal.Decimal
			switch tt.op {
			case "add":
				result = d1.Add(d2)
			case "sub":
				result = d1.Sub(d2)
			case "mul":
				result = d1.Mul(d2)
			case "div":
				result = d1.Div(d2).Round(8) // Round to 8 decimal places
			}

			// For addition, format to 8 decimal places for comparison
			if tt.op == "add" || tt.op == "sub" {
				result = result.Round(8)
				resultStr := result.StringFixed(8)
				if resultStr != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, resultStr)
				}
			} else {
				if result.String() != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result.String())
				}
			}
		})
	}
}

func TestDecimalValidation(t *testing.T) {
	// Test decimal validation for trading amounts
	tests := []struct {
		name    string
		amount  string
		valid   bool
		scale   int32
	}{
		{
			name:   "Valid BTC amount",
			amount: "0.12345678",
			valid:  true,
			scale:  8,
		},
		{
			name:   "Valid USDT amount",
			amount: "1234.56",
			valid:  true,
			scale:  2,
		},
		{
			name:   "Invalid precision - too many decimals",
			amount: "0.123456789",
			valid:  false,
			scale:  8,
		},
		{
			name:   "Zero amount",
			amount: "0",
			valid:  true,
			scale:  8,
		},
		{
			name:   "Negative amount",
			amount: "-1.23",
			valid:  false,
			scale:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := decimal.NewFromString(tt.amount)
			if err != nil {
				if tt.valid {
					t.Errorf("Expected valid decimal, got error: %v", err)
				}
				return
			}

			// Check if amount is negative
			if d.IsNegative() && tt.valid {
				t.Errorf("Expected positive amount, got negative: %s", tt.amount)
				return
			}

			// Check decimal places
			if d.Exponent() < -tt.scale && tt.valid {
				t.Errorf("Expected max %d decimal places, got %d", tt.scale, -d.Exponent())
				return
			}

			if !tt.valid && d.Exponent() >= -tt.scale && !d.IsNegative() {
				t.Errorf("Expected invalid decimal, but validation passed for: %s", tt.amount)
			}
		})
	}
}

// ValidateAmount validates trading amounts with proper precision
func ValidateAmount(amount string, scale int32) error {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		return ErrInvalidAmount
	}

	if d.IsNegative() {
		return ErrInvalidAmount
	}

	if d.Exponent() < -scale {
		return ErrInvalidAmount
	}

	return nil
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name      string
		amount    string
		scale     int32
		expectErr bool
	}{
		{
			name:      "Valid BTC amount",
			amount:    "0.12345678",
			scale:     8,
			expectErr: false,
		},
		{
			name:      "Invalid precision",
			amount:    "0.123456789",
			scale:     8,
			expectErr: true,
		},
		{
			name:      "Negative amount",
			amount:    "-1.23",
			scale:     2,
			expectErr: true,
		},
		{
			name:      "Invalid format",
			amount:    "abc",
			scale:     8,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount, tt.scale)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}