package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TradingPairModel = (*customTradingPairModel)(nil)

type (
	// TradingPairModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTradingPairModel.
	TradingPairModel interface {
		tradingPairModel
		// 自定义方法
		FindBySymbol(ctx context.Context, symbol string) (*TradingPair, error)
		FindByStatus(ctx context.Context, status int64) ([]*TradingPair, error)
		FindActivePairs(ctx context.Context) ([]*TradingPair, error)
	}

	customTradingPairModel struct {
		*defaultTradingPairModel
	}

	// TradingPair 交易对配置模型
	TradingPair struct {
		ID            uint64    `db:"id"`             // 交易对ID，主键
		Symbol        string    `db:"symbol"`         // 交易对符号，格式：基础币种/计价币种，如BTC/USDT
		BaseCurrency  string    `db:"base_currency"`  // 基础币种代码，交易的目标币种
		QuoteCurrency string    `db:"quote_currency"` // 计价币种代码，用于定价的币种
		MinAmount     string    `db:"min_amount"`     // 单笔交易最小数量限制
		MaxAmount     string    `db:"max_amount"`     // 单笔交易最大数量限制
		PriceScale    int64     `db:"price_scale"`    // 价格显示精度，小数点后位数
		AmountScale   int64     `db:"amount_scale"`   // 数量显示精度，小数点后位数
		Status        int64     `db:"status"`         // 交易对状态：1-正常交易，2-禁用交易
		CreatedAt     time.Time `db:"created_at"`     // 交易对创建时间
	}

	tradingPairModel interface {
		Insert(ctx context.Context, data *TradingPair) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*TradingPair, error)
		Update(ctx context.Context, data *TradingPair) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultTradingPairModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewTradingPairModel returns a model for the database table.
func NewTradingPairModel(conn sqlx.SqlConn) TradingPairModel {
	return &customTradingPairModel{
		defaultTradingPairModel: newTradingPairModel(conn),
	}
}

func newTradingPairModel(conn sqlx.SqlConn) *defaultTradingPairModel {
	return &defaultTradingPairModel{
		conn:  conn,
		table: "trading_pairs",
	}
}

func (m *defaultTradingPairModel) Insert(ctx context.Context, data *TradingPair) (sql.Result, error) {
	query := `INSERT INTO ` + m.table + ` (symbol, base_currency, quote_currency, min_amount, max_amount, price_scale, amount_scale, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	ret, err := m.conn.ExecCtx(ctx, query, data.Symbol, data.BaseCurrency, data.QuoteCurrency, data.MinAmount, data.MaxAmount, data.PriceScale, data.AmountScale, data.Status, data.CreatedAt)
	return ret, err
}

func (m *defaultTradingPairModel) FindOne(ctx context.Context, id uint64) (*TradingPair, error) {
	query := `SELECT id, symbol, base_currency, quote_currency, min_amount, max_amount, price_scale, amount_scale, status, created_at FROM ` + m.table + ` WHERE id = $1 LIMIT 1`
	var resp TradingPair
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customTradingPairModel) FindBySymbol(ctx context.Context, symbol string) (*TradingPair, error) {
	query := `SELECT id, symbol, base_currency, quote_currency, min_amount, max_amount, price_scale, amount_scale, status, created_at FROM ` + m.table + ` WHERE symbol = $1 LIMIT 1`
	var resp TradingPair
	err := m.conn.QueryRowCtx(ctx, &resp, query, symbol)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customTradingPairModel) FindByStatus(ctx context.Context, status int64) ([]*TradingPair, error) {
	query := `SELECT id, symbol, base_currency, quote_currency, min_amount, max_amount, price_scale, amount_scale, status, created_at FROM ` + m.table + ` WHERE status = $1`
	var resp []*TradingPair
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	return resp, err
}

func (m *customTradingPairModel) FindActivePairs(ctx context.Context) ([]*TradingPair, error) {
	query := `SELECT id, symbol, base_currency, quote_currency, min_amount, max_amount, price_scale, amount_scale, status, created_at FROM ` + m.table + ` WHERE status = 1`
	var resp []*TradingPair
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *defaultTradingPairModel) Update(ctx context.Context, data *TradingPair) error {
	query := `UPDATE ` + m.table + ` SET symbol = $1, base_currency = $2, quote_currency = $3, min_amount = $4, max_amount = $5, price_scale = $6, amount_scale = $7, status = $8 WHERE id = $9`
	_, err := m.conn.ExecCtx(ctx, query, data.Symbol, data.BaseCurrency, data.QuoteCurrency, data.MinAmount, data.MaxAmount, data.PriceScale, data.AmountScale, data.Status, data.ID)
	return err
}

func (m *defaultTradingPairModel) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM ` + m.table + ` WHERE id = $1`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}