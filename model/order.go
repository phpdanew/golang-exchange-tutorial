package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OrderModel = (*customOrderModel)(nil)

type (
	// OrderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOrderModel.
	OrderModel interface {
		orderModel
		// 自定义方法
		FindByUserID(ctx context.Context, userID uint64) ([]*Order, error)
		FindBySymbol(ctx context.Context, symbol string) ([]*Order, error)
		FindByStatus(ctx context.Context, status int64) ([]*Order, error)
		FindByUserIDAndStatus(ctx context.Context, userID uint64, status int64) ([]*Order, error)
		FindBySymbolAndStatus(ctx context.Context, symbol string, status int64) ([]*Order, error)
		FindBySymbolAndSideAndStatus(ctx context.Context, symbol string, side int64, status int64) ([]*Order, error)
		// 分页查询方法
		FindByUserIDWithPagination(ctx context.Context, userID uint64, symbol string, status int64, page, size int64) ([]*Order, int64, error)
		UpdateStatus(ctx context.Context, id uint64, status int64) error
		UpdateFilledAmount(ctx context.Context, id uint64, filledAmount string) error
		Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error
	}

	customOrderModel struct {
		*defaultOrderModel
	}

	// Order 交易订单模型
	Order struct {
		ID           uint64    `db:"id"`            // 订单ID，主键
		UserID       uint64    `db:"user_id"`       // 下单用户ID，关联users表
		Symbol       string    `db:"symbol"`        // 交易对符号，如BTC/USDT
		Type         int64     `db:"type"`          // 订单类型：1-限价单（指定价格），2-市价单（按市场价格）
		Side         int64     `db:"side"`          // 交易方向：1-买入（买入基础币种），2-卖出（卖出基础币种）
		Amount       string    `db:"amount"`        // 订单总数量，基础币种数量
		Price        string    `db:"price"`         // 订单价格，限价单必填，市价单为NULL
		FilledAmount string    `db:"filled_amount"` // 已成交数量，累计成交的基础币种数量
		Status       int64     `db:"status"`        // 订单状态：1-待成交，2-部分成交，3-完全成交，4-已取消
		CreatedAt    time.Time `db:"created_at"`    // 订单创建时间
		UpdatedAt    time.Time `db:"updated_at"`    // 订单最后更新时间
	}

	orderModel interface {
		Insert(ctx context.Context, data *Order) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*Order, error)
		Update(ctx context.Context, data *Order) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultOrderModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewOrderModel returns a model for the database table.
func NewOrderModel(conn sqlx.SqlConn) OrderModel {
	return &customOrderModel{
		defaultOrderModel: newOrderModel(conn),
	}
}

func newOrderModel(conn sqlx.SqlConn) *defaultOrderModel {
	return &defaultOrderModel{
		conn:  conn,
		table: "orders",
	}
}

func (m *defaultOrderModel) Insert(ctx context.Context, data *Order) (sql.Result, error) {
	query := `INSERT INTO ` + m.table + ` (user_id, symbol, type, side, amount, price, filled_amount, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	ret, err := m.conn.ExecCtx(ctx, query, data.UserID, data.Symbol, data.Type, data.Side, data.Amount, data.Price, data.FilledAmount, data.Status, data.CreatedAt, data.UpdatedAt)
	return ret, err
}

func (m *defaultOrderModel) FindOne(ctx context.Context, id uint64) (*Order, error) {
	query := `SELECT id, user_id, symbol, type, side, amount, price, filled_amount, status, created_at, updated_at FROM ` + m.table + ` WHERE id = $1 LIMIT 1`
	var resp Order
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

func (m *customOrderModel) FindByUserID(ctx context.Context, userID uint64) ([]*Order, error) {
	query := `SELECT id, user_id, symbol, type, side, amount, price, filled_amount, status, created_at, updated_at FROM ` + m.table + ` WHERE user_id = $1 ORDER BY created_at DESC`
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID)
	return resp, err
}

func (m *customOrderModel) FindBySymbol(ctx context.Context, symbol string) ([]*Order, error) {
	query := `SELECT id, user_id, symbol, type, side, amount, price, filled_amount, status, created_at, updated_at FROM ` + m.table + ` WHERE symbol = $1 ORDER BY created_at DESC`
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, symbol)
	return resp, err
}

func (m *customOrderModel) FindByStatus(ctx context.Context, status int64) ([]*Order, error) {
	query := `SELECT id, user_id, symbol, type, side, amount, price, filled_amount, status, created_at, updated_at FROM ` + m.table + ` WHERE status = $1 ORDER BY created_at DESC`
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	return resp, err
}

func (m *customOrderModel) FindByUserIDAndStatus(ctx context.Context, userID uint64, status int64) ([]*Order, error) {
	query := `SELECT id, user_id, symbol, type, side, amount, price, filled_amount, status, created_at, updated_at FROM ` + m.table + ` WHERE user_id = $1 AND status = $2 ORDER BY created_at DESC`
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID, status)
	return resp, err
}

func (m *customOrderModel) FindBySymbolAndStatus(ctx context.Context, symbol string, status int64) ([]*Order, error) {
	query := `SELECT id, user_id, symbol, type, side, amount, price, filled_amount, status, created_at, updated_at FROM ` + m.table + ` WHERE symbol = $1 AND status = $2 ORDER BY created_at DESC`
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, symbol, status)
	return resp, err
}

func (m *customOrderModel) FindBySymbolAndSideAndStatus(ctx context.Context, symbol string, side int64, status int64) ([]*Order, error) {
	query := `SELECT id, user_id, symbol, type, side, amount, price, filled_amount, status, created_at, updated_at FROM ` + m.table + ` WHERE symbol = $1 AND side = $2 AND status = $3 ORDER BY price ASC, created_at ASC`
	var resp []*Order
	err := m.conn.QueryRowsCtx(ctx, &resp, query, symbol, side, status)
	return resp, err
}

func (m *customOrderModel) UpdateStatus(ctx context.Context, id uint64, status int64) error {
	query := `UPDATE ` + m.table + ` SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := m.conn.ExecCtx(ctx, query, status, time.Now(), id)
	return err
}

func (m *customOrderModel) UpdateFilledAmount(ctx context.Context, id uint64, filledAmount string) error {
	query := `UPDATE ` + m.table + ` SET filled_amount = $1, updated_at = $2 WHERE id = $3`
	_, err := m.conn.ExecCtx(ctx, query, filledAmount, time.Now(), id)
	return err
}

func (m *defaultOrderModel) Update(ctx context.Context, data *Order) error {
	query := `UPDATE ` + m.table + ` SET user_id = $1, symbol = $2, type = $3, side = $4, amount = $5, price = $6, filled_amount = $7, status = $8, updated_at = $9 WHERE id = $10`
	_, err := m.conn.ExecCtx(ctx, query, data.UserID, data.Symbol, data.Type, data.Side, data.Amount, data.Price, data.FilledAmount, data.Status, data.UpdatedAt, data.ID)
	return err
}

func (m *defaultOrderModel) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM ` + m.table + ` WHERE id = $1`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *customOrderModel) Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, fn)
}

// FindByUserIDWithPagination 分页查询用户订单
func (m *customOrderModel) FindByUserIDWithPagination(ctx context.Context, userID uint64, symbol string, status int64, page, size int64) ([]*Order, int64, error) {
	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100 // 限制最大页面大小
	}

	offset := (page - 1) * size

	// 构建查询条件和参数
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	paramIndex := 2

	if symbol != "" {
		whereClause += " AND symbol = $2"
		args = append(args, symbol)
		paramIndex = 3
	}

	if status > 0 {
		if symbol != "" {
			whereClause += " AND status = $3"
		} else {
			whereClause += " AND status = $2"
		}
		args = append(args, status)
		paramIndex++
	}

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM ` + m.table + ` ` + whereClause
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据 - 构建LIMIT和OFFSET参数
	limitClause := ""
	if symbol != "" && status > 0 {
		limitClause = " ORDER BY created_at DESC LIMIT $4 OFFSET $5"
	} else if symbol != "" || status > 0 {
		limitClause = " ORDER BY created_at DESC LIMIT $3 OFFSET $4"
	} else {
		limitClause = " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	}

	dataQuery := `SELECT id, user_id, symbol, type, side, amount, price, filled_amount, status, created_at, updated_at FROM ` + m.table + ` ` + whereClause + limitClause
	args = append(args, size, offset)

	var resp []*Order
	err = m.conn.QueryRowsCtx(ctx, &resp, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}