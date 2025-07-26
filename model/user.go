package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		// 自定义方法
		FindOneByEmail(ctx context.Context, email string) (*User, error)
	}

	customUserModel struct {
		*defaultUserModel
	}

	// User 用户基础信息模型
	User struct {
		ID        uint64    `db:"id"`         // 用户ID，主键
		Email     string    `db:"email"`      // 用户邮箱，唯一标识
		Password  string    `db:"password"`   // 密码哈希值，使用bcrypt加密
		Nickname  string    `db:"nickname"`   // 用户昵称，显示名称
		Status    int64     `db:"status"`     // 用户状态：1-正常，2-禁用，3-删除
		CreatedAt time.Time `db:"created_at"` // 账户创建时间
		UpdatedAt time.Time `db:"updated_at"` // 最后更新时间
	}

	userModel interface {
		Insert(ctx context.Context, data *User) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*User, error)
		Update(ctx context.Context, data *User) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultUserModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn),
	}
}

func newUserModel(conn sqlx.SqlConn) *defaultUserModel {
	return &defaultUserModel{
		conn:  conn,
		table: "users",
	}
}

func (m *defaultUserModel) Insert(ctx context.Context, data *User) (sql.Result, error) {
	query := `INSERT INTO ` + m.table + ` (email, password, nickname, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	ret, err := m.conn.ExecCtx(ctx, query, data.Email, data.Password, data.Nickname, data.Status, data.CreatedAt, data.UpdatedAt)
	return ret, err
}

func (m *defaultUserModel) FindOne(ctx context.Context, id uint64) (*User, error) {
	query := `SELECT id, email, password, nickname, status, created_at, updated_at FROM ` + m.table + ` WHERE id = $1 LIMIT 1`
	var resp User
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

func (m *customUserModel) FindOneByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, password, nickname, status, created_at, updated_at FROM ` + m.table + ` WHERE email = $1 LIMIT 1`
	var resp User
	err := m.conn.QueryRowCtx(ctx, &resp, query, email)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultUserModel) Update(ctx context.Context, data *User) error {
	query := `UPDATE ` + m.table + ` SET email = $1, password = $2, nickname = $3, status = $4, updated_at = $5 WHERE id = $6`
	_, err := m.conn.ExecCtx(ctx, query, data.Email, data.Password, data.Nickname, data.Status, data.UpdatedAt, data.ID)
	return err
}

func (m *defaultUserModel) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM ` + m.table + ` WHERE id = $1`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}