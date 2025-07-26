package asset

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DepositLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDepositLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DepositLogic {
	return &DepositLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DepositLogic) Deposit(req *types.DepositRequest) (resp *types.DepositResponse, err error) {
	// 1. 从JWT上下文中获取用户ID
	userID, err := l.getUserIDFromContext()
	if err != nil {
		l.Errorf("Failed to get user ID from context: %v", err)
		return nil, model.ErrUnauthorized
	}

	// 2. 验证充值参数
	if err := l.validateDepositRequest(req); err != nil {
		l.Errorf("Invalid deposit request for user %d: %v", userID, err)
		return nil, err
	}

	// 3. 解析充值金额
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		l.Errorf("Invalid amount format for user %d: %s", userID, req.Amount)
		return nil, model.ErrInvalidAmount
	}

	// 4. 生成交易ID
	transactionID := l.generateTransactionID()

	// 5. 使用数据库事务处理充值
	err = l.svcCtx.BalanceModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 查找或创建用户余额记录
		balance, err := l.svcCtx.BalanceModel.FindByUserIDAndCurrency(ctx, userID, req.Currency)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				// 创建新的余额记录
				newBalance := &model.Balance{
					UserID:    userID,
					Currency:  req.Currency,
					Available: req.Amount, // 充值金额直接加到可用余额
					Frozen:    "0.00000000",
					UpdatedAt: time.Now(),
				}
				_, err = l.svcCtx.BalanceModel.Insert(ctx, newBalance)
				if err != nil {
					l.Errorf("Failed to create balance record for user %d, currency %s: %v", userID, req.Currency, err)
					return err // 返回原始错误用于测试
				}
			} else {
				l.Errorf("Failed to find balance for user %d, currency %s: %v", userID, req.Currency, err)
				return model.ErrInternalServer
			}
		} else {
			// 更新现有余额记录
			currentAvailable, err := decimal.NewFromString(balance.Available)
			if err != nil {
				l.Errorf("Invalid current available balance format for user %d, currency %s: %s", userID, req.Currency, balance.Available)
				return model.ErrInternalServer
			}

			// 计算新的可用余额
			newAvailable := currentAvailable.Add(amount)

			// 更新余额
			err = l.svcCtx.BalanceModel.UpdateBalance(ctx, userID, req.Currency, newAvailable.String(), balance.Frozen)
			if err != nil {
				l.Errorf("Failed to update balance for user %d, currency %s: %v", userID, req.Currency, err)
				return err // 返回原始错误用于测试
			}
		}

		// 创建交易记录
		now := time.Now()
		transaction := &model.AssetTransaction{
			UserID:        userID,
			TransactionID: transactionID,
			Currency:      req.Currency,
			Type:          1, // 1-充值
			Amount:        req.Amount,
			Fee:           "0.00000000", // 充值通常不收手续费
			Status:        2,            // 2-成功（在真实场景中，这里应该是1-待处理，等待区块链确认后再更新为成功）
			Address:       "",           // 充值地址可以为空，或者从请求中获取
			TxHash:        "",           // 区块链交易哈希，实际场景中需要从区块链获取
			Remark:        fmt.Sprintf("Deposit %s %s", req.Amount, req.Currency),
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		_, err = l.svcCtx.AssetTransactionModel.Insert(ctx, transaction)
		if err != nil {
			l.Errorf("Failed to create transaction record for user %d, transaction ID %s: %v", userID, transactionID, err)
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 6. 构造响应
	resp = &types.DepositResponse{
		TransactionID: transactionID,
		Currency:      req.Currency,
		Amount:        req.Amount,
		Status:        2, // 2-成功（在真实场景中，这里应该是1-处理中，等待区块链确认后再更新为成功）
		CreatedAt:     time.Now().Format("2006-01-02 15:04:05"),
	}

	l.Infof("Deposit successful for user %d: %s %s, transaction ID: %s", userID, req.Amount, req.Currency, transactionID)
	return resp, nil
}

// getUserIDFromContext 从JWT上下文中获取用户ID
func (l *DepositLogic) getUserIDFromContext() (uint64, error) {
	userIDValue := l.ctx.Value("userId")
	if userIDValue == nil {
		return 0, errors.New("user ID not found in context")
	}

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

// validateDepositRequest 验证充值请求参数
func (l *DepositLogic) validateDepositRequest(req *types.DepositRequest) error {
	// 验证币种代码
	if req.Currency == "" {
		return model.ErrInvalidParams
	}

	// 币种代码应该是大写字母
	req.Currency = strings.ToUpper(req.Currency)

	// 验证支持的币种（这里简化处理，实际应该从配置或数据库读取）
	supportedCurrencies := map[string]bool{
		"BTC":  true,
		"ETH":  true,
		"USDT": true,
		"USDC": true,
	}

	if !supportedCurrencies[req.Currency] {
		return model.ErrCurrencyNotFound
	}

	// 验证充值金额
	if req.Amount == "" {
		return model.ErrInvalidAmount
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return model.ErrInvalidAmount
	}

	// 充值金额必须大于0
	if amount.LessThanOrEqual(decimal.Zero) {
		return model.ErrInvalidAmount
	}

	// 验证金额精度（最多8位小数）
	if amount.Exponent() < -8 {
		return errors.New("amount precision exceeds 8 decimal places")
	}

	// 验证最小充值金额（这里设置一个通用的最小值）
	minAmount := decimal.NewFromFloat(0.00000001) // 0.00000001
	if amount.LessThan(minAmount) {
		return errors.New("amount below minimum deposit limit")
	}

	// 验证最大充值金额（防止异常大额充值）
	maxAmount := decimal.NewFromInt(1000000) // 1,000,000
	if amount.GreaterThan(maxAmount) {
		return errors.New("amount exceeds maximum deposit limit")
	}

	return nil
}

// generateTransactionID 生成唯一的交易ID
func (l *DepositLogic) generateTransactionID() string {
	// 使用UUID生成唯一ID，并添加前缀标识这是充值交易
	return fmt.Sprintf("DEP_%s", strings.ReplaceAll(uuid.New().String(), "-", ""))
}
