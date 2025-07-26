package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KlineModel = (*customKlineModel)(nil)

type (
	// KlineModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKlineModel.
	KlineModel interface {
		klineModel
		// 自定义方法
		FindBySymbolAndInterval(ctx context.Context, symbol, interval string) ([]*Kline, error)
		FindBySymbolAndIntervalWithLimit(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error)
		FindBySymbolAndIntervalAndTimeRange(ctx context.Context, symbol, interval string, startTime, endTime time.Time) ([]*Kline, error)
		FindLatestBySymbolAndInterval(ctx context.Context, symbol, interval string) (*Kline, error)
		UpsertKline(ctx context.Context, data *Kline) error
		Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error
	}

	customKlineModel struct {
		*defaultKlineModel
	}

	// Kline K线数据模型，存储各时间周期的OHLCV数据
	Kline struct {
		ID        uint64    `db:"id"`         // K线记录ID，主键
		Symbol    string    `db:"symbol"`     // 交易对符号，如BTC/USDT
		Interval  string    `db:"interval"`   // 时间周期，如1m(1分钟)、5m(5分钟)、1h(1小时)、1d(1天)等
		OpenTime  time.Time `db:"open_time"`  // K线周期开始时间
		CloseTime time.Time `db:"close_time"` // K线周期结束时间
		Open      string    `db:"open"`       // 开盘价，周期内第一笔成交价格
		High      string    `db:"high"`       // 最高价，周期内最高成交价格
		Low       string    `db:"low"`        // 最低价，周期内最低成交价格
		Close     string    `db:"close"`      // 收盘价，周期内最后一笔成交价格
		Volume    string    `db:"volume"`     // 成交量，周期内累计成交的基础币种数量
	}

	klineModel interface {
		Insert(ctx context.Context, data *Kline) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*Kline, error)
		Update(ctx context.Context, data *Kline) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultKlineModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewKlineModel returns a model for the database table.
func NewKlineModel(conn sqlx.SqlConn) KlineModel {
	return &customKlineModel{
		defaultKlineModel: newKlineModel(conn),
	}
}

func newKlineModel(conn sqlx.SqlConn) *defaultKlineModel {
	return &defaultKlineModel{
		conn:  conn,
		table: "klines",
	}
}

func (m *defaultKlineModel) Insert(ctx context.Context, data *Kline) (sql.Result, error) {
	query := `INSERT INTO ` + m.table + ` (symbol, interval, open_time, close_time, open, high, low, close, volume) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	ret, err := m.conn.ExecCtx(ctx, query, data.Symbol, data.Interval, data.OpenTime, data.CloseTime, data.Open, data.High, data.Low, data.Close, data.Volume)
	return ret, err
}

func (m *defaultKlineModel) FindOne(ctx context.Context, id uint64) (*Kline, error) {
	query := `SELECT id, symbol, interval, open_time, close_time, open, high, low, close, volume FROM ` + m.table + ` WHERE id = $1 LIMIT 1`
	var resp Kline
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

func (m *customKlineModel) FindBySymbolAndInterval(ctx context.Context, symbol, interval string) ([]*Kline, error) {
	query := `SELECT id, symbol, interval, open_time, close_time, open, high, low, close, volume FROM ` + m.table + ` WHERE symbol = $1 AND interval = $2 ORDER BY open_time DESC`
	var resp []*Kline
	err := m.conn.QueryRowsCtx(ctx, &resp, query, symbol, interval)
	return resp, err
}

func (m *customKlineModel) FindBySymbolAndIntervalWithLimit(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error) {
	query := `SELECT id, symbol, interval, open_time, close_time, open, high, low, close, volume FROM ` + m.table + ` WHERE symbol = $1 AND interval = $2 ORDER BY open_time DESC LIMIT $3`
	var resp []*Kline
	err := m.conn.QueryRowsCtx(ctx, &resp, query, symbol, interval, limit)
	return resp, err
}

func (m *customKlineModel) FindBySymbolAndIntervalAndTimeRange(ctx context.Context, symbol, interval string, startTime, endTime time.Time) ([]*Kline, error) {
	query := `SELECT id, symbol, interval, open_time, close_time, open, high, low, close, volume FROM ` + m.table + ` WHERE symbol = $1 AND interval = $2 AND open_time >= $3 AND open_time <= $4 ORDER BY open_time ASC`
	var resp []*Kline
	err := m.conn.QueryRowsCtx(ctx, &resp, query, symbol, interval, startTime, endTime)
	return resp, err
}

func (m *customKlineModel) FindLatestBySymbolAndInterval(ctx context.Context, symbol, interval string) (*Kline, error) {
	query := `SELECT id, symbol, interval, open_time, close_time, open, high, low, close, volume FROM ` + m.table + ` WHERE symbol = $1 AND interval = $2 ORDER BY open_time DESC LIMIT 1`
	var resp Kline
	err := m.conn.QueryRowCtx(ctx, &resp, query, symbol, interval)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customKlineModel) UpsertKline(ctx context.Context, data *Kline) error {
	query := `INSERT INTO ` + m.table + ` (symbol, interval, open_time, close_time, open, high, low, close, volume) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			  ON CONFLICT (symbol, interval, open_time) 
			  DO UPDATE SET close_time = $4, open = $5, high = $6, low = $7, close = $8, volume = $9`
	_, err := m.conn.ExecCtx(ctx, query, data.Symbol, data.Interval, data.OpenTime, data.CloseTime, data.Open, data.High, data.Low, data.Close, data.Volume)
	return err
}

func (m *defaultKlineModel) Update(ctx context.Context, data *Kline) error {
	query := `UPDATE ` + m.table + ` SET symbol = $1, interval = $2, open_time = $3, close_time = $4, open = $5, high = $6, low = $7, close = $8, volume = $9 WHERE id = $10`
	_, err := m.conn.ExecCtx(ctx, query, data.Symbol, data.Interval, data.OpenTime, data.CloseTime, data.Open, data.High, data.Low, data.Close, data.Volume, data.ID)
	return err
}

func (m *defaultKlineModel) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM ` + m.table + ` WHERE id = $1`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *customKlineModel) Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, fn)
}