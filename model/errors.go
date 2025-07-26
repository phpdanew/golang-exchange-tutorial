package model

import "errors"

// 通用错误 / Common Errors
var (
	ErrNotFound        = errors.New("record not found")
	ErrInternalServer  = errors.New("internal server error")
	ErrInvalidParams   = errors.New("invalid parameters")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
)

// 用户相关错误 / User Related Errors
var (
	ErrUserExists      = errors.New("user already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrUserDisabled    = errors.New("user account is disabled")
)

// 资产相关错误 / Asset Related Errors
var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrCurrencyNotFound    = errors.New("currency not found")
	ErrBalanceNotFound     = errors.New("balance not found")
)

// 交易相关错误 / Trading Related Errors
var (
	ErrInvalidOrderType    = errors.New("invalid order type")
	ErrInvalidOrderSide    = errors.New("invalid order side")
	ErrOrderNotFound       = errors.New("order not found")
	ErrTradingPairNotFound = errors.New("trading pair not found")
	ErrTradingPairDisabled = errors.New("trading pair is disabled")
	ErrOrderAlreadyCanceled = errors.New("order already canceled")
	ErrOrderAlreadyFilled   = errors.New("order already filled")
)

// 市场数据相关错误 / Market Data Related Errors
var (
	ErrInvalidTimeRange = errors.New("invalid time range")
	ErrInvalidInterval  = errors.New("invalid interval")
	ErrNoMarketData     = errors.New("no market data available")
)