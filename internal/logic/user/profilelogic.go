package user

import (
	"context"
	"errors"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileLogic {
	return &ProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProfileLogic) Profile() (resp *types.User, err error) {
	// 1. 从JWT上下文中获取用户ID
	userID, err := l.getUserIDFromContext()
	if err != nil {
		l.Errorf("Failed to get user ID from context: %v", err)
		return nil, model.ErrUnauthorized
	}

	// 2. 从数据库查询用户信息
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, model.ErrUserNotFound
		}
		l.Errorf("Failed to find user by ID: %v", err)
		return nil, model.ErrInternalServer
	}

	// 3. 检查用户状态
	if user.Status != 1 {
		return nil, model.ErrUserDisabled
	}

	// 4. 返回用户信息（不包含密码）
	resp = &types.User{
		ID:       user.ID,
		Email:    user.Email,
		Nickname: user.Nickname,
		Status:   user.Status,
	}

	l.Infof("User profile retrieved successfully: %d", userID)
	return resp, nil
}

// getUserIDFromContext 从JWT上下文中获取用户ID
func (l *ProfileLogic) getUserIDFromContext() (uint64, error) {
	// 在go-zero中，JWT中间件会将解析后的claims存储在context中
	// 键名通常是 "userId"，这与我们在登录时设置的claims一致
	userIDValue := l.ctx.Value("userId")
	if userIDValue == nil {
		return 0, errors.New("user ID not found in context")
	}

	// JWT claims中的数值通常是float64类型
	switch v := userIDValue.(type) {
	case float64:
		return uint64(v), nil
	case uint64:
		return v, nil
	case int64:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	default:
		l.Errorf("Invalid user ID type in context: %T", v)
		return 0, errors.New("invalid user ID type in context")
	}
}
