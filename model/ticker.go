package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TickerModel = (*customTickerModel)(nil)

type (
	// TickerModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTickerModel.
	TickerModel interface {
		tickerModel
		// 自定义方法
		FindBySymbol(ctx context.Context, symbol string) (*Ticker, error)
		FindAll(ctx context.Context) ([]*Ticker, error)
		UpsertTicker(ctx context.Context, data *Ticker) error
		UpdatePrice(ctx context.Context, symbol, price string) error
		UpdateStats(ctx context.Context, symbol, lastPrice, change24h, volume24h, high24h, low24h string) error
		Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error
	}

	customTickerModel struct {
		*defaultTickerModel
	}

	// Ticker 24小时市场统计数据模型
	Ticker struct {
		Symbol    string    `db:"symbol"`     // 交易对符号，如BTC/USDT，主键
		LastPrice string    `db:"last_price"` // 最新成交价格
		Change24h string    `db:"change_24h"` // 24小时价格变化金额（当前价格-24小时前价格）
		Volume24h string    `db:"volume_24h"` // 24小时累计成交量，基础币种数量
		High24h   string    `db:"high_24h"`   // 24小时内最高成交价格
		Low24h    string    `db:"low_24h"`    // 24小时内最低成交价格
		UpdatedAt time.Time `db:"updated_at"` // 数据最后更新时间
	}

	tickerModel interface {
		Insert(ctx context.Context, data *Ticker) (sql.Result, error)
		FindOne(ctx context.Context, symbol string) (*Ticker, error)
		Update(ctx context.Context, data *Ticker) error
		Delete(ctx context.Context, symbol string) error
	}

	defaultTickerModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewTickerModel returns a model for the database table.
func NewTickerModel(conn sqlx.SqlConn) TickerModel {
	return &customTickerModel{
		defaultTickerModel: newTickerModel(conn),
	}
}

func newTickerModel(conn sqlx.SqlConn) *defaultTickerModel {
	return &defaultTickerModel{
		conn:  conn,
		table: "tickers",
	}
}

func (m *defaultTickerModel) Insert(ctx context.Context, data *Ticker) (sql.Result, error) {
	query := `INSERT INTO ` + m.table + ` (symbol, last_price, change_24h, volume_24h, high_24h, low_24h, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	ret, err := m.conn.ExecCtx(ctx, query, data.Symbol, data.LastPrice, data.Change24h, data.Volume24h, data.High24h, data.Low24h, data.UpdatedAt)
	return ret, err
}

func (m *defaultTickerModel) FindOne(ctx context.Context, symbol string) (*Ticker, error) {
	query := `SELECT symbol, last_price, change_24h, volume_24h, high_24h, low_24h, updated_at FROM ` + m.table + ` WHERE symbol = $1 LIMIT 1`
	var resp Ticker
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

func (m *customTickerModel) FindBySymbol(ctx context.Context, symbol string) (*Ticker, error) {
	return m.FindOne(ctx, symbol)
}

func (m *customTickerModel) FindAll(ctx context.Context) ([]*Ticker, error) {
	query := `SELECT symbol, last_price, change_24h, volume_24h, high_24h, low_24h, updated_at FROM ` + m.table + ` ORDER BY symbol`
	var resp []*Ticker
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *customTickerModel) UpsertTicker(ctx context.Context, data *Ticker) error {
	query := `INSERT INTO ` + m.table + ` (symbol, last_price, change_24h, volume_24h, high_24h, low_24h, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)
			  ON CONFLICT (symbol) 
			  DO UPDATE SET last_price = $2, change_24h = $3, volume_24h = $4, high_24h = $5, low_24h = $6, updated_at = $7`
	_, err := m.conn.ExecCtx(ctx, query, data.Symbol, data.LastPrice, data.Change24h, data.Volume24h, data.High24h, data.Low24h, data.UpdatedAt)
	return err
}

func (m *customTickerModel) UpdatePrice(ctx context.Context, symbol, price string) error {
	query := `UPDATE ` + m.table + ` SET last_price = $1, updated_at = $2 WHERE symbol = $3`
	_, err := m.conn.ExecCtx(ctx, query, price, time.Now(), symbol)
	return err
}

func (m *customTickerModel) UpdateStats(ctx context.Context, symbol, lastPrice, change24h, volume24h, high24h, low24h string) error {
	query := `UPDATE ` + m.table + ` SET last_price = $1, change_24h = $2, volume_24h = $3, high_24h = $4, low_24h = $5, updated_at = $6 WHERE symbol = $7`
	_, err := m.conn.ExecCtx(ctx, query, lastPrice, change24h, volume24h, high24h, low24h, time.Now(), symbol)
	return err
}

func (m *defaultTickerModel) Update(ctx context.Context, data *Ticker) error {
	query := `UPDATE ` + m.table + ` SET last_price = $1, change_24h = $2, volume_24h = $3, high_24h = $4, low_24h = $5, updated_at = $6 WHERE symbol = $7`
	_, err := m.conn.ExecCtx(ctx, query, data.LastPrice, data.Change24h, data.Volume24h, data.High24h, data.Low24h, data.UpdatedAt, data.Symbol)
	return err
}

func (m *defaultTickerModel) Delete(ctx context.Context, symbol string) error {
	query := `DELETE FROM ` + m.table + ` WHERE symbol = $1`
	_, err := m.conn.ExecCtx(ctx, query, symbol)
	return err
}

func (m *customTickerModel) Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, fn)
}