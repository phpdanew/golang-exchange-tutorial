package model

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BalanceModel = (*customBalanceModel)(nil)

type (
	// BalanceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBalanceModel.
	BalanceModel interface {
		balanceModel
		// 自定义方法
		FindByUserID(ctx context.Context, userID uint64) ([]*Balance, error)
		FindByUserIDAndCurrency(ctx context.Context, userID uint64, currency string) (*Balance, error)
		UpdateBalance(ctx context.Context, userID uint64, currency string, available, frozen string) error
		FreezeBalance(ctx context.Context, userID uint64, currency string, amount string) error
		UnfreezeBalance(ctx context.Context, userID uint64, currency string, amount string) error
		Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error
	}

	customBalanceModel struct {
		*defaultBalanceModel
	}

	// Balance 用户资产余额模型
	Balance struct {
		ID        uint64    `db:"id"`         // 余额记录ID，主键
		UserID    uint64    `db:"user_id"`    // 用户ID，关联users表
		Currency  string    `db:"currency"`   // 币种代码，如BTC、ETH、USDT等
		Available string    `db:"available"`  // 可用余额，使用string存储decimal避免精度问题
		Frozen    string    `db:"frozen"`     // 冻结余额，挂单时冻结的金额
		UpdatedAt time.Time `db:"updated_at"` // 余额最后更新时间
	}

	balanceModel interface {
		Insert(ctx context.Context, data *Balance) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*Balance, error)
		Update(ctx context.Context, data *Balance) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultBalanceModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewBalanceModel returns a model for the database table.
func NewBalanceModel(conn sqlx.SqlConn) BalanceModel {
	return &customBalanceModel{
		defaultBalanceModel: newBalanceModel(conn),
	}
}

func newBalanceModel(conn sqlx.SqlConn) *defaultBalanceModel {
	return &defaultBalanceModel{
		conn:  conn,
		table: "balances",
	}
}

func (m *defaultBalanceModel) Insert(ctx context.Context, data *Balance) (sql.Result, error) {
	query := `INSERT INTO ` + m.table + ` (user_id, currency, available, frozen, updated_at) VALUES ($1, $2, $3, $4, $5)`
	ret, err := m.conn.ExecCtx(ctx, query, data.UserID, data.Currency, data.Available, data.Frozen, data.UpdatedAt)
	return ret, err
}

func (m *defaultBalanceModel) FindOne(ctx context.Context, id uint64) (*Balance, error) {
	query := `SELECT id, user_id, currency, available, frozen, updated_at FROM ` + m.table + ` WHERE id = $1 LIMIT 1`
	var resp Balance
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

func (m *customBalanceModel) FindByUserID(ctx context.Context, userID uint64) ([]*Balance, error) {
	query := `SELECT id, user_id, currency, available, frozen, updated_at FROM ` + m.table + ` WHERE user_id = $1`
	var resp []*Balance
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID)
	return resp, err
}

func (m *customBalanceModel) FindByUserIDAndCurrency(ctx context.Context, userID uint64, currency string) (*Balance, error) {
	query := `SELECT id, user_id, currency, available, frozen, updated_at FROM ` + m.table + ` WHERE user_id = $1 AND currency = $2 LIMIT 1`
	var resp Balance
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID, currency)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customBalanceModel) UpdateBalance(ctx context.Context, userID uint64, currency string, available, frozen string) error {
	query := `UPDATE ` + m.table + ` SET available = $1, frozen = $2, updated_at = $3 WHERE user_id = $4 AND currency = $5`
	_, err := m.conn.ExecCtx(ctx, query, available, frozen, time.Now(), userID, currency)
	return err
}

func (m *defaultBalanceModel) Update(ctx context.Context, data *Balance) error {
	query := `UPDATE ` + m.table + ` SET user_id = $1, currency = $2, available = $3, frozen = $4, updated_at = $5 WHERE id = $6`
	_, err := m.conn.ExecCtx(ctx, query, data.UserID, data.Currency, data.Available, data.Frozen, data.UpdatedAt, data.ID)
	return err
}

func (m *defaultBalanceModel) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM ` + m.table + ` WHERE id = $1`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *customBalanceModel) FreezeBalance(ctx context.Context, userID uint64, currency string, amount string) error {
	// 使用decimal进行精确计算
	freezeAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return errors.New("invalid freeze amount format")
	}

	if freezeAmount.LessThanOrEqual(decimal.Zero) {
		return errors.New("freeze amount must be positive")
	}

	// 查询当前余额
	balance, err := m.FindByUserIDAndCurrency(ctx, userID, currency)
	if err != nil {
		return err
	}

	currentAvailable, err := decimal.NewFromString(balance.Available)
	if err != nil {
		return errors.New("invalid current available balance format")
	}

	currentFrozen, err := decimal.NewFromString(balance.Frozen)
	if err != nil {
		return errors.New("invalid current frozen balance format")
	}

	// 检查可用余额是否充足
	if currentAvailable.LessThan(freezeAmount) {
		return ErrInsufficientBalance
	}

	// 计算新的余额
	newAvailable := currentAvailable.Sub(freezeAmount)
	newFrozen := currentFrozen.Add(freezeAmount)

	// 更新余额
	return m.UpdateBalance(ctx, userID, currency, newAvailable.String(), newFrozen.String())
}

func (m *customBalanceModel) UnfreezeBalance(ctx context.Context, userID uint64, currency string, amount string) error {
	// 使用decimal进行精确计算
	unfreezeAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return errors.New("invalid unfreeze amount format")
	}

	if unfreezeAmount.LessThanOrEqual(decimal.Zero) {
		return errors.New("unfreeze amount must be positive")
	}

	// 查询当前余额
	balance, err := m.FindByUserIDAndCurrency(ctx, userID, currency)
	if err != nil {
		return err
	}

	currentAvailable, err := decimal.NewFromString(balance.Available)
	if err != nil {
		return errors.New("invalid current available balance format")
	}

	currentFrozen, err := decimal.NewFromString(balance.Frozen)
	if err != nil {
		return errors.New("invalid current frozen balance format")
	}

	// 检查冻结余额是否充足
	if currentFrozen.LessThan(unfreezeAmount) {
		return errors.New("insufficient frozen balance")
	}

	// 计算新的余额
	newAvailable := currentAvailable.Add(unfreezeAmount)
	newFrozen := currentFrozen.Sub(unfreezeAmount)

	// 更新余额
	return m.UpdateBalance(ctx, userID, currency, newAvailable.String(), newFrozen.String())
}

func (m *customBalanceModel) Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, fn)
}