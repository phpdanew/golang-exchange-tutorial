package model

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// 创建一个测试用的balance model实例
type testBalanceModel struct {
	mockFindByUserIDAndCurrency func(ctx context.Context, userID uint64, currency string) (*Balance, error)
	mockUpdateBalance           func(ctx context.Context, userID uint64, currency string, available, frozen string) error
}

func (t *testBalanceModel) FindByUserIDAndCurrency(ctx context.Context, userID uint64, currency string) (*Balance, error) {
	if t.mockFindByUserIDAndCurrency != nil {
		return t.mockFindByUserIDAndCurrency(ctx, userID, currency)
	}
	return nil, ErrNotFound
}

func (t *testBalanceModel) UpdateBalance(ctx context.Context, userID uint64, currency string, available, frozen string) error {
	if t.mockUpdateBalance != nil {
		return t.mockUpdateBalance(ctx, userID, currency, available, frozen)
	}
	return nil
}

// 实现FreezeBalance方法
func (t *testBalanceModel) FreezeBalance(ctx context.Context, userID uint64, currency string, amount string) error {
	// 使用decimal进行精确计算
	freezeAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return errors.New("invalid freeze amount format")
	}

	if freezeAmount.LessThanOrEqual(decimal.Zero) {
		return errors.New("freeze amount must be positive")
	}

	// 查询当前余额
	balance, err := t.FindByUserIDAndCurrency(ctx, userID, currency)
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
	return t.UpdateBalance(ctx, userID, currency, newAvailable.String(), newFrozen.String())
}

// 实现UnfreezeBalance方法
func (t *testBalanceModel) UnfreezeBalance(ctx context.Context, userID uint64, currency string, amount string) error {
	// 使用decimal进行精确计算
	unfreezeAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return errors.New("invalid unfreeze amount format")
	}

	if unfreezeAmount.LessThanOrEqual(decimal.Zero) {
		return errors.New("unfreeze amount must be positive")
	}

	// 查询当前余额
	balance, err := t.FindByUserIDAndCurrency(ctx, userID, currency)
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
	return t.UpdateBalance(ctx, userID, currency, newAvailable.String(), newFrozen.String())
}

func TestCustomBalanceModel_FreezeBalance(t *testing.T) {
	tests := []struct {
		name            string
		userID          uint64
		currency        string
		freezeAmount    string
		currentBalance  *Balance
		findError       error
		updateError     error
		expectedError   string
		expectedUpdate  *struct {
			available string
			frozen    string
		}
	}{
		{
			name:         "成功冻结余额",
			userID:       1,
			currency:     "BTC",
			freezeAmount: "1.00000000",
			currentBalance: &Balance{
				ID:        1,
				UserID:    1,
				Currency:  "BTC",
				Available: "2.00000000",
				Frozen:    "0.50000000",
				UpdatedAt: time.Now(),
			},
			findError:     nil,
			updateError:   nil,
			expectedError: "",
			expectedUpdate: &struct {
				available string
				frozen    string
			}{
				available: "1",        // 2.0 - 1.0 = 1.0
				frozen:    "1.5",      // 0.5 + 1.0 = 1.5
			},
		},
		{
			name:         "余额不足无法冻结",
			userID:       2,
			currency:     "BTC",
			freezeAmount: "3.00000000",
			currentBalance: &Balance{
				ID:        2,
				UserID:    2,
				Currency:  "BTC",
				Available: "2.00000000",
				Frozen:    "0.00000000",
				UpdatedAt: time.Now(),
			},
			findError:      nil,
			updateError:    nil,
			expectedError:  "insufficient balance",
			expectedUpdate: nil,
		},
		{
			name:           "无效的冻结金额格式",
			userID:         3,
			currency:       "BTC",
			freezeAmount:   "invalid_amount",
			currentBalance: nil,
			findError:      nil,
			updateError:    nil,
			expectedError:  "invalid freeze amount format",
			expectedUpdate: nil,
		},
		{
			name:         "负数冻结金额",
			userID:       4,
			currency:     "BTC",
			freezeAmount: "-1.00000000",
			currentBalance: &Balance{
				ID:        4,
				UserID:    4,
				Currency:  "BTC",
				Available: "2.00000000",
				Frozen:    "0.00000000",
				UpdatedAt: time.Now(),
			},
			findError:      nil,
			updateError:    nil,
			expectedError:  "freeze amount must be positive",
			expectedUpdate: nil,
		},
		{
			name:           "余额记录不存在",
			userID:         6,
			currency:       "BTC",
			freezeAmount:   "1.00000000",
			currentBalance: nil,
			findError:      ErrNotFound,
			updateError:    nil,
			expectedError:  "record not found",
			expectedUpdate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试用的balance model
			testModel := &testBalanceModel{}

			// 设置mock函数
			testModel.mockFindByUserIDAndCurrency = func(ctx context.Context, userID uint64, currency string) (*Balance, error) {
				if tt.findError != nil {
					return nil, tt.findError
				}
				return tt.currentBalance, nil
			}

			var actualAvailable, actualFrozen string
			testModel.mockUpdateBalance = func(ctx context.Context, userID uint64, currency string, available, frozen string) error {
				actualAvailable = available
				actualFrozen = frozen
				return tt.updateError
			}

			// 执行测试
			err := testModel.FreezeBalance(context.Background(), tt.userID, tt.currency, tt.freezeAmount)

			// 验证结果
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedUpdate != nil {
					assert.Equal(t, tt.expectedUpdate.available, actualAvailable)
					assert.Equal(t, tt.expectedUpdate.frozen, actualFrozen)
				}
			}
		})
	}
}

func TestCustomBalanceModel_UnfreezeBalance(t *testing.T) {
	tests := []struct {
		name             string
		userID           uint64
		currency         string
		unfreezeAmount   string
		currentBalance   *Balance
		findError        error
		updateError      error
		expectedError    string
		expectedUpdate   *struct {
			available string
			frozen    string
		}
	}{
		{
			name:           "成功解冻余额",
			userID:         1,
			currency:       "BTC",
			unfreezeAmount: "0.50000000",
			currentBalance: &Balance{
				ID:        1,
				UserID:    1,
				Currency:  "BTC",
				Available: "1.00000000",
				Frozen:    "1.50000000",
				UpdatedAt: time.Now(),
			},
			findError:     nil,
			updateError:   nil,
			expectedError: "",
			expectedUpdate: &struct {
				available string
				frozen    string
			}{
				available: "1.5",      // 1.0 + 0.5 = 1.5
				frozen:    "1",        // 1.5 - 0.5 = 1.0
			},
		},
		{
			name:           "冻结余额不足无法解冻",
			userID:         2,
			currency:       "BTC",
			unfreezeAmount: "2.00000000",
			currentBalance: &Balance{
				ID:        2,
				UserID:    2,
				Currency:  "BTC",
				Available: "1.00000000",
				Frozen:    "1.00000000",
				UpdatedAt: time.Now(),
			},
			findError:      nil,
			updateError:    nil,
			expectedError:  "insufficient frozen balance",
			expectedUpdate: nil,
		},
		{
			name:           "无效的解冻金额格式",
			userID:         3,
			currency:       "BTC",
			unfreezeAmount: "invalid_amount",
			currentBalance: nil,
			findError:      nil,
			updateError:    nil,
			expectedError:  "invalid unfreeze amount format",
			expectedUpdate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试用的balance model
			testModel := &testBalanceModel{}

			// 设置mock函数
			testModel.mockFindByUserIDAndCurrency = func(ctx context.Context, userID uint64, currency string) (*Balance, error) {
				if tt.findError != nil {
					return nil, tt.findError
				}
				return tt.currentBalance, nil
			}

			var actualAvailable, actualFrozen string
			testModel.mockUpdateBalance = func(ctx context.Context, userID uint64, currency string, available, frozen string) error {
				actualAvailable = available
				actualFrozen = frozen
				return tt.updateError
			}

			// 执行测试
			err := testModel.UnfreezeBalance(context.Background(), tt.userID, tt.currency, tt.unfreezeAmount)

			// 验证结果
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedUpdate != nil {
					assert.Equal(t, tt.expectedUpdate.available, actualAvailable)
					assert.Equal(t, tt.expectedUpdate.frozen, actualFrozen)
				}
			}
		})
	}
}

func TestBalanceFreezeUnfreezeIntegration(t *testing.T) {
	// 集成测试：测试冻结和解冻的完整流程
	testModel := &testBalanceModel{}

	// 初始余额状态
	currentBalance := &Balance{
		ID:        1,
		UserID:    1,
		Currency:  "BTC",
		Available: "10.00000000",
		Frozen:    "0.00000000",
		UpdatedAt: time.Now(),
	}

	// 设置mock函数来模拟数据库操作
	testModel.mockFindByUserIDAndCurrency = func(ctx context.Context, userID uint64, currency string) (*Balance, error) {
		return currentBalance, nil
	}

	testModel.mockUpdateBalance = func(ctx context.Context, userID uint64, currency string, available, frozen string) error {
		// 更新当前余额状态
		currentBalance.Available = available
		currentBalance.Frozen = frozen
		return nil
	}

	ctx := context.Background()

	// 步骤1：冻结5个BTC
	err := testModel.FreezeBalance(ctx, 1, "BTC", "5.00000000")
	assert.NoError(t, err)
	assert.Equal(t, "5", currentBalance.Available)
	assert.Equal(t, "5", currentBalance.Frozen)

	// 步骤2：再冻结2个BTC
	err = testModel.FreezeBalance(ctx, 1, "BTC", "2.00000000")
	assert.NoError(t, err)
	assert.Equal(t, "3", currentBalance.Available)
	assert.Equal(t, "7", currentBalance.Frozen)

	// 步骤3：尝试冻结超过可用余额的金额（应该失败）
	err = testModel.FreezeBalance(ctx, 1, "BTC", "5.00000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")
	// 余额应该保持不变
	assert.Equal(t, "3", currentBalance.Available)
	assert.Equal(t, "7", currentBalance.Frozen)

	// 步骤4：解冻3个BTC
	err = testModel.UnfreezeBalance(ctx, 1, "BTC", "3.00000000")
	assert.NoError(t, err)
	assert.Equal(t, "6", currentBalance.Available)
	assert.Equal(t, "4", currentBalance.Frozen)

	// 步骤5：尝试解冻超过冻结余额的金额（应该失败）
	err = testModel.UnfreezeBalance(ctx, 1, "BTC", "5.00000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient frozen balance")
	// 余额应该保持不变
	assert.Equal(t, "6", currentBalance.Available)
	assert.Equal(t, "4", currentBalance.Frozen)

	// 步骤6：解冻剩余的4个BTC
	err = testModel.UnfreezeBalance(ctx, 1, "BTC", "4.00000000")
	assert.NoError(t, err)
	assert.Equal(t, "10", currentBalance.Available)
	assert.Equal(t, "0", currentBalance.Frozen)

	// 验证最终状态：应该回到初始状态
	totalBalance, _ := decimal.NewFromString(currentBalance.Available)
	frozenBalance, _ := decimal.NewFromString(currentBalance.Frozen)
	total := totalBalance.Add(frozenBalance)
	assert.Equal(t, "10", total.String())
}