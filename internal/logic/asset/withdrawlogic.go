package asset

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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

type WithdrawLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawLogic {
	return &WithdrawLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawLogic) Withdraw(req *types.WithdrawRequest) (resp *types.WithdrawResponse, err error) {
	// 1. 从JWT上下文中获取用户ID
	userID, err := l.getUserIDFromContext()
	if err != nil {
		l.Errorf("Failed to get user ID from context: %v", err)
		return nil, model.ErrUnauthorized
	}

	// 2. 验证提现参数
	if err := l.validateWithdrawRequest(req); err != nil {
		l.Errorf("Invalid withdraw request for user %d: %v", userID, err)
		return nil, err
	}

	// 3. 解析提现金额
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		l.Errorf("Invalid amount format for user %d: %s", userID, req.Amount)
		return nil, model.ErrInvalidAmount
	}

	// 4. 计算提现手续费
	fee := l.calculateWithdrawFee(req.Currency, amount)
	totalAmount := amount.Add(fee) // 总扣除金额 = 提现金额 + 手续费

	// 5. 生成交易ID
	transactionID := l.generateTransactionID()

	// 6. 使用数据库事务处理提现
	err = l.svcCtx.BalanceModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 查找用户余额记录
		balance, err := l.svcCtx.BalanceModel.FindByUserIDAndCurrency(ctx, userID, req.Currency)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				l.Errorf("Balance not found for user %d, currency %s", userID, req.Currency)
				return model.ErrBalanceNotFound
			}
			l.Errorf("Failed to find balance for user %d, currency %s: %v", userID, req.Currency, err)
			return err
		}

		// 检查余额充足性
		currentAvailable, err := decimal.NewFromString(balance.Available)
		if err != nil {
			l.Errorf("Invalid current available balance format for user %d, currency %s: %s", userID, req.Currency, balance.Available)
			return model.ErrInternalServer
		}

		if currentAvailable.LessThan(totalAmount) {
			l.Errorf("Insufficient balance for user %d: available %s, required %s (including fee %s)", 
				userID, currentAvailable.String(), totalAmount.String(), fee.String())
			return model.ErrInsufficientBalance
		}

		// 扣减余额
		newAvailable := currentAvailable.Sub(totalAmount)

		// 更新余额
		err = l.svcCtx.BalanceModel.UpdateBalance(ctx, userID, req.Currency, newAvailable.String(), balance.Frozen)
		if err != nil {
			l.Errorf("Failed to update balance for user %d, currency %s: %v", userID, req.Currency, err)
			return err
		}

		// 创建交易记录
		now := time.Now()
		transaction := &model.AssetTransaction{
			UserID:        userID,
			TransactionID: transactionID,
			Currency:      req.Currency,
			Type:          2, // 2-提现
			Amount:        req.Amount,
			Fee:           fee.String(),
			Status:        1,           // 1-待审核（提现通常需要人工审核）
			Address:       req.Address, // 提现地址
			TxHash:        "",          // 区块链交易哈希，实际场景中在区块链确认后更新
			Remark:        fmt.Sprintf("Withdraw %s %s to %s", req.Amount, req.Currency, req.Address),
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

	// 7. 构造响应
	resp = &types.WithdrawResponse{
		TransactionID: transactionID,
		Currency:      req.Currency,
		Amount:        req.Amount,
		Address:       req.Address,
		Fee:           fee.String(),
		Status:        1, // 1-待审核（在真实场景中，提现通常需要人工审核）
		CreatedAt:     time.Now().Format("2006-01-02 15:04:05"),
	}

	l.Infof("Withdraw request created for user %d: %s %s to %s, fee: %s, transaction ID: %s", 
		userID, req.Amount, req.Currency, req.Address, fee.String(), transactionID)
	return resp, nil
}

// getUserIDFromContext 从JWT上下文中获取用户ID
func (l *WithdrawLogic) getUserIDFromContext() (uint64, error) {
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

// validateWithdrawRequest 验证提现请求参数
func (l *WithdrawLogic) validateWithdrawRequest(req *types.WithdrawRequest) error {
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

	// 验证提现金额
	if req.Amount == "" {
		return model.ErrInvalidAmount
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return model.ErrInvalidAmount
	}

	// 提现金额必须大于0
	if amount.LessThanOrEqual(decimal.Zero) {
		return model.ErrInvalidAmount
	}

	// 验证金额精度（最多8位小数）
	if amount.Exponent() < -8 {
		return errors.New("amount precision exceeds 8 decimal places")
	}

	// 验证最小提现金额
	minAmount := l.getMinWithdrawAmount(req.Currency)
	if amount.LessThan(minAmount) {
		return fmt.Errorf("amount below minimum withdraw limit: %s", minAmount.String())
	}

	// 验证最大提现金额
	maxAmount := l.getMaxWithdrawAmount(req.Currency)
	if amount.GreaterThan(maxAmount) {
		return fmt.Errorf("amount exceeds maximum withdraw limit: %s", maxAmount.String())
	}

	// 验证提现地址
	if req.Address == "" {
		return errors.New("withdraw address is required")
	}

	if err := l.validateWithdrawAddress(req.Currency, req.Address); err != nil {
		return err
	}

	return nil
}

// validateWithdrawAddress 验证提现地址格式
func (l *WithdrawLogic) validateWithdrawAddress(currency, address string) error {
	// 简化的地址验证，实际应该根据不同币种使用不同的验证规则
	switch currency {
	case "BTC":
		// Bitcoin地址验证（简化版）
		if len(address) < 26 || len(address) > 35 {
			return errors.New("invalid Bitcoin address format")
		}
		// 检查是否以1、3或bc1开头
		if !strings.HasPrefix(address, "1") && !strings.HasPrefix(address, "3") && !strings.HasPrefix(address, "bc1") {
			return errors.New("invalid Bitcoin address format")
		}
	case "ETH", "USDT", "USDC":
		// Ethereum地址验证（简化版）
		if len(address) != 42 {
			return errors.New("invalid Ethereum address format")
		}
		if !strings.HasPrefix(address, "0x") {
			return errors.New("invalid Ethereum address format")
		}
		// 检查是否为有效的十六进制字符
		matched, _ := regexp.MatchString("^0x[a-fA-F0-9]{40}$", address)
		if !matched {
			return errors.New("invalid Ethereum address format")
		}
	default:
		return errors.New("unsupported currency for address validation")
	}

	return nil
}

// calculateWithdrawFee 计算提现手续费
func (l *WithdrawLogic) calculateWithdrawFee(currency string, amount decimal.Decimal) decimal.Decimal {
	// 简化的手续费计算，实际应该从配置或数据库读取
	switch currency {
	case "BTC":
		// BTC固定手续费 0.0005
		return decimal.NewFromFloat(0.0005)
	case "ETH":
		// ETH固定手续费 0.005
		return decimal.NewFromFloat(0.005)
	case "USDT", "USDC":
		// USDT/USDC固定手续费 1.0
		return decimal.NewFromFloat(1.0)
	default:
		// 默认手续费为提现金额的0.1%
		return amount.Mul(decimal.NewFromFloat(0.001))
	}
}

// getMinWithdrawAmount 获取最小提现金额
func (l *WithdrawLogic) getMinWithdrawAmount(currency string) decimal.Decimal {
	// 简化处理，实际应该从配置读取
	switch currency {
	case "BTC":
		return decimal.NewFromFloat(0.001)
	case "ETH":
		return decimal.NewFromFloat(0.01)
	case "USDT", "USDC":
		return decimal.NewFromFloat(10.0)
	default:
		return decimal.NewFromFloat(0.00000001)
	}
}

// getMaxWithdrawAmount 获取最大提现金额
func (l *WithdrawLogic) getMaxWithdrawAmount(currency string) decimal.Decimal {
	// 简化处理，实际应该从配置读取
	switch currency {
	case "BTC":
		return decimal.NewFromFloat(10.0)
	case "ETH":
		return decimal.NewFromFloat(100.0)
	case "USDT", "USDC":
		return decimal.NewFromFloat(100000.0)
	default:
		return decimal.NewFromFloat(1000000.0)
	}
}

// generateTransactionID 生成唯一的交易ID
func (l *WithdrawLogic) generateTransactionID() string {
	// 使用UUID生成唯一ID，并添加前缀标识这是提现交易
	return fmt.Sprintf("WTH_%s", strings.ReplaceAll(uuid.New().String(), "-", ""))
}
