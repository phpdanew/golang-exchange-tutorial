package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TradeModel = (*customTradeModel)(nil)

type (
	// TradeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTradeModel.
	TradeModel interface {
		tradeModel
		// 自定义方法
		FindBySymbol(ctx context.Context, symbol string) ([]*Trade, error)
		FindBySymbolWithLimit(ctx context.Context, symbol string, limit int) ([]*Trade, error)
		FindByUserID(ctx context.Context, userID uint64) ([]*Trade, error)
		FindByOrderID(ctx context.Context, orderID uint64) ([]*Trade, error)
		FindByTimeRange(ctx context.Context, symbol string, startTime, endTime time.Time) ([]*Trade, error)
		Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error
	}

	customTradeModel struct {
		*defaultTradeModel
	}

	// Trade 交易成交记录模型
	Trade struct {
		ID          uint64    `db:"id"`            // 成交记录ID，主键
		Symbol      string    `db:"symbol"`        // 交易对符号，如BTC/USDT
		BuyOrderID  uint64    `db:"buy_order_id"`  // 买单订单ID，关联orders表
		SellOrderID uint64    `db:"sell_order_id"` // 卖单订单ID，关联orders表
		BuyUserID   uint64    `db:"buy_user_id"`   // 买方用户ID，关联users表
		SellUserID  uint64    `db:"sell_user_id"`  // 卖方用户ID，关联users表
		Price       string    `db:"price"`         // 成交价格，以计价币种计价
		Amount      string    `db:"amount"`        // 成交数量，基础币种数量
		CreatedAt   time.Time `db:"created_at"`    // 成交时间戳
	}

	tradeModel interface {
		Insert(ctx context.Context, data *Trade) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*Trade, error)
		Update(ctx context.Context, data *Trade) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultTradeModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewTradeModel returns a model for the database table.
func NewTradeModel(conn sqlx.SqlConn) TradeModel {
	return &customTradeModel{
		defaultTradeModel: newTradeModel(conn),
	}
}

func newTradeModel(conn sqlx.SqlConn) *defaultTradeModel {
	return &defaultTradeModel{
		conn:  conn,
		table: "trades",
	}
}

func (m *defaultTradeModel) Insert(ctx context.Context, data *Trade) (sql.Result, error) {
	query := `INSERT INTO ` + m.table + ` (symbol, buy_order_id, sell_order_id, buy_user_id, sell_user_id, price, amount, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	ret, err := m.conn.ExecCtx(ctx, query, data.Symbol, data.BuyOrderID, data.SellOrderID, data.BuyUserID, data.SellUserID, data.Price, data.Amount, data.CreatedAt)
	return ret, err
}

func (m *defaultTradeModel) FindOne(ctx context.Context, id uint64) (*Trade, error) {
	query := `SELECT id, symbol, buy_order_id, sell_order_id, buy_user_id, sell_user_id, price, amount, created_at FROM ` + m.table + ` WHERE id = $1 LIMIT 1`
	var resp Trade
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

func (m *customTradeModel) FindBySymbol(ctx context.Context, symbol string) ([]*Trade, error) {
	query := `SELECT id, symbol, buy_order_id, sell_order_id, buy_user_id, sell_user_id, price, amount, created_at FROM ` + m.table + ` WHERE symbol = $1 ORDER BY created_at DESC`
	var resp []*Trade
	err := m.conn.QueryRowsCtx(ctx, &resp, query, symbol)
	return resp, err
}

func (m *customTradeModel) FindBySymbolWithLimit(ctx context.Context, symbol string, limit int) ([]*Trade, error) {
	query := `SELECT id, symbol, buy_order_id, sell_order_id, buy_user_id, sell_user_id, price, amount, created_at FROM ` + m.table + ` WHERE symbol = $1 ORDER BY created_at DESC LIMIT $2`
	var resp []*Trade
	err := m.conn.QueryRowsCtx(ctx, &resp, query, symbol, limit)
	return resp, err
}

func (m *customTradeModel) FindByUserID(ctx context.Context, userID uint64) ([]*Trade, error) {
	query := `SELECT id, symbol, buy_order_id, sell_order_id, buy_user_id, sell_user_id, price, amount, created_at FROM ` + m.table + ` WHERE buy_user_id = $1 OR sell_user_id = $1 ORDER BY created_at DESC`
	var resp []*Trade
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID)
	return resp, err
}

func (m *customTradeModel) FindByOrderID(ctx context.Context, orderID uint64) ([]*Trade, error) {
	query := `SELECT id, symbol, buy_order_id, sell_order_id, buy_user_id, sell_user_id, price, amount, created_at FROM ` + m.table + ` WHERE buy_order_id = $1 OR sell_order_id = $1 ORDER BY created_at DESC`
	var resp []*Trade
	err := m.conn.QueryRowsCtx(ctx, &resp, query, orderID)
	return resp, err
}

func (m *customTradeModel) FindByTimeRange(ctx context.Context, symbol string, startTime, endTime time.Time) ([]*Trade, error) {
	query := `SELECT id, symbol, buy_order_id, sell_order_id, buy_user_id, sell_user_id, price, amount, created_at FROM ` + m.table + ` WHERE symbol = $1 AND created_at >= $2 AND created_at <= $3 ORDER BY created_at DESC`
	var resp []*Trade
	err := m.conn.QueryRowsCtx(ctx, &resp, query, symbol, startTime, endTime)
	return resp, err
}

func (m *defaultTradeModel) Update(ctx context.Context, data *Trade) error {
	query := `UPDATE ` + m.table + ` SET symbol = $1, buy_order_id = $2, sell_order_id = $3, buy_user_id = $4, sell_user_id = $5, price = $6, amount = $7 WHERE id = $8`
	_, err := m.conn.ExecCtx(ctx, query, data.Symbol, data.BuyOrderID, data.SellOrderID, data.BuyUserID, data.SellUserID, data.Price, data.Amount, data.ID)
	return err
}

func (m *defaultTradeModel) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM ` + m.table + ` WHERE id = $1`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *customTradeModel) Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, fn)
}