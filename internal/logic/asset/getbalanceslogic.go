package asset

import (
	"context"
	"errors"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetBalancesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBalancesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBalancesLogic {
	return &GetBalancesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBalancesLogic) GetBalances() (resp *types.BalanceResponse, err error) {
	// 1. 从JWT上下文中获取用户ID
	userID, err := l.getUserIDFromContext()
	if err != nil {
		l.Errorf("Failed to get user ID from context: %v", err)
		return nil, model.ErrUnauthorized
	}

	// 2. 从数据库查询用户所有币种余额
	balances, err := l.svcCtx.BalanceModel.FindByUserID(l.ctx, userID)
	if err != nil {
		l.Errorf("Failed to find balances for user %d: %v", userID, err)
		return nil, model.ErrInternalServer
	}

	// 3. 转换为响应格式，确保数值精度
	var responseBalances []types.Balance
	for _, balance := range balances {
		// 验证数值格式的有效性
		if err := l.validateBalanceAmounts(balance.Available, balance.Frozen); err != nil {
			l.Errorf("Invalid balance amounts for user %d, currency %s: %v", userID, balance.Currency, err)
			continue // 跳过无效的余额记录，但不中断整个查询
		}

		responseBalances = append(responseBalances, types.Balance{
			Currency:  balance.Currency,
			Available: balance.Available, // 保持string格式确保精度
			Frozen:    balance.Frozen,    // 保持string格式确保精度
			UpdatedAt: balance.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// 4. 如果用户没有任何余额记录，返回空列表而不是错误
	if len(responseBalances) == 0 {
		l.Infof("No balance records found for user %d", userID)
		responseBalances = []types.Balance{} // 确保返回空数组而不是nil
	}

	resp = &types.BalanceResponse{
		Balances: responseBalances,
	}

	l.Infof("Successfully retrieved %d balance records for user %d", len(responseBalances), userID)
	return resp, nil
}

// getUserIDFromContext 从JWT上下文中获取用户ID
func (l *GetBalancesLogic) getUserIDFromContext() (uint64, error) {
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

// validateBalanceAmounts 验证余额数值的有效性
func (l *GetBalancesLogic) validateBalanceAmounts(available, frozen string) error {
	// 验证可用余额格式
	availableDecimal, err := decimal.NewFromString(available)
	if err != nil {
		return errors.New("invalid available balance format")
	}

	// 验证冻结余额格式
	frozenDecimal, err := decimal.NewFromString(frozen)
	if err != nil {
		return errors.New("invalid frozen balance format")
	}

	// 验证余额不能为负数
	if availableDecimal.IsNegative() {
		return errors.New("available balance cannot be negative")
	}

	if frozenDecimal.IsNegative() {
		return errors.New("frozen balance cannot be negative")
	}

	return nil
}
