package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AssetTransactionModel = (*customAssetTransactionModel)(nil)

type (
	// AssetTransactionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAssetTransactionModel.
	AssetTransactionModel interface {
		assetTransactionModel
		// 自定义方法
		FindByUserID(ctx context.Context, userID uint64, limit, offset int64) ([]*AssetTransaction, error)
		FindByTransactionID(ctx context.Context, transactionID string) (*AssetTransaction, error)
		FindByUserIDAndType(ctx context.Context, userID uint64, transactionType int64, limit, offset int64) ([]*AssetTransaction, error)
		CountByUserID(ctx context.Context, userID uint64) (int64, error)
		CountByUserIDAndType(ctx context.Context, userID uint64, transactionType int64) (int64, error)
	}

	customAssetTransactionModel struct {
		*defaultAssetTransactionModel
	}

	// AssetTransaction 资产交易记录模型
	AssetTransaction struct {
		ID            uint64    `db:"id"`             // 交易记录ID，主键
		UserID        uint64    `db:"user_id"`        // 用户ID，关联users表
		TransactionID string    `db:"transaction_id"` // 交易ID，唯一标识
		Currency      string    `db:"currency"`       // 币种代码，如BTC、ETH、USDT等
		Type          int64     `db:"type"`           // 交易类型：1-充值，2-提现
		Amount        string    `db:"amount"`         // 交易金额，使用string存储decimal避免精度问题
		Fee           string    `db:"fee"`            // 手续费，使用string存储decimal避免精度问题
		Status        int64     `db:"status"`         // 交易状态：1-待处理，2-成功，3-失败，4-已取消
		Address       string    `db:"address"`        // 地址（提现时有值，充值时可为空）
		TxHash        string    `db:"tx_hash"`        // 区块链交易哈希（可选）
		Remark        string    `db:"remark"`         // 备注信息
		CreatedAt     time.Time `db:"created_at"`     // 创建时间
		UpdatedAt     time.Time `db:"updated_at"`     // 更新时间
	}

	assetTransactionModel interface {
		Insert(ctx context.Context, data *AssetTransaction) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*AssetTransaction, error)
		Update(ctx context.Context, data *AssetTransaction) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultAssetTransactionModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewAssetTransactionModel returns a model for the database table.
func NewAssetTransactionModel(conn sqlx.SqlConn) AssetTransactionModel {
	return &customAssetTransactionModel{
		defaultAssetTransactionModel: newAssetTransactionModel(conn),
	}
}

func newAssetTransactionModel(conn sqlx.SqlConn) *defaultAssetTransactionModel {
	return &defaultAssetTransactionModel{
		conn:  conn,
		table: "asset_transactions",
	}
}

func (m *defaultAssetTransactionModel) Insert(ctx context.Context, data *AssetTransaction) (sql.Result, error) {
	query := `INSERT INTO ` + m.table + ` (user_id, transaction_id, currency, type, amount, fee, status, address, tx_hash, remark, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	ret, err := m.conn.ExecCtx(ctx, query, data.UserID, data.TransactionID, data.Currency, data.Type, data.Amount, data.Fee, data.Status, data.Address, data.TxHash, data.Remark, data.CreatedAt, data.UpdatedAt)
	return ret, err
}

func (m *defaultAssetTransactionModel) FindOne(ctx context.Context, id uint64) (*AssetTransaction, error) {
	query := `SELECT id, user_id, transaction_id, currency, type, amount, fee, status, address, tx_hash, remark, created_at, updated_at FROM ` + m.table + ` WHERE id = $1 LIMIT 1`
	var resp AssetTransaction
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

func (m *customAssetTransactionModel) FindByTransactionID(ctx context.Context, transactionID string) (*AssetTransaction, error) {
	query := `SELECT id, user_id, transaction_id, currency, type, amount, fee, status, address, tx_hash, remark, created_at, updated_at FROM ` + m.table + ` WHERE transaction_id = $1 LIMIT 1`
	var resp AssetTransaction
	err := m.conn.QueryRowCtx(ctx, &resp, query, transactionID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAssetTransactionModel) FindByUserID(ctx context.Context, userID uint64, limit, offset int64) ([]*AssetTransaction, error) {
	query := `SELECT id, user_id, transaction_id, currency, type, amount, fee, status, address, tx_hash, remark, created_at, updated_at FROM ` + m.table + ` WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	var resp []*AssetTransaction
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID, limit, offset)
	return resp, err
}

func (m *customAssetTransactionModel) FindByUserIDAndType(ctx context.Context, userID uint64, transactionType int64, limit, offset int64) ([]*AssetTransaction, error) {
	query := `SELECT id, user_id, transaction_id, currency, type, amount, fee, status, address, tx_hash, remark, created_at, updated_at FROM ` + m.table + ` WHERE user_id = $1 AND type = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	var resp []*AssetTransaction
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID, transactionType, limit, offset)
	return resp, err
}

func (m *customAssetTransactionModel) CountByUserID(ctx context.Context, userID uint64) (int64, error) {
	query := `SELECT COUNT(*) FROM ` + m.table + ` WHERE user_id = $1`
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, userID)
	return count, err
}

func (m *customAssetTransactionModel) CountByUserIDAndType(ctx context.Context, userID uint64, transactionType int64) (int64, error) {
	query := `SELECT COUNT(*) FROM ` + m.table + ` WHERE user_id = $1 AND type = $2`
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, userID, transactionType)
	return count, err
}

func (m *defaultAssetTransactionModel) Update(ctx context.Context, data *AssetTransaction) error {
	query := `UPDATE ` + m.table + ` SET user_id = $1, transaction_id = $2, currency = $3, type = $4, amount = $5, fee = $6, status = $7, address = $8, tx_hash = $9, remark = $10, updated_at = $11 WHERE id = $12`
	_, err := m.conn.ExecCtx(ctx, query, data.UserID, data.TransactionID, data.Currency, data.Type, data.Amount, data.Fee, data.Status, data.Address, data.TxHash, data.Remark, data.UpdatedAt, data.ID)
	return err
}

func (m *defaultAssetTransactionModel) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM ` + m.table + ` WHERE id = $1`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}