package model

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockAssetTransactionModel 用于测试的模拟实现
type MockAssetTransactionModel struct {
	transactions map[string]*AssetTransaction
	nextID       uint64
}

func NewMockAssetTransactionModel() *MockAssetTransactionModel {
	return &MockAssetTransactionModel{
		transactions: make(map[string]*AssetTransaction),
		nextID:       1,
	}
}

func (m *MockAssetTransactionModel) Insert(ctx context.Context, data *AssetTransaction) (interface{}, error) {
	// 模拟插入操作
	data.ID = m.nextID
	m.nextID++
	m.transactions[data.TransactionID] = data
	return nil, nil
}

func (m *MockAssetTransactionModel) FindByTransactionID(ctx context.Context, transactionID string) (*AssetTransaction, error) {
	if tx, exists := m.transactions[transactionID]; exists {
		return tx, nil
	}
	return nil, ErrNotFound
}

func (m *MockAssetTransactionModel) FindByUserID(ctx context.Context, userID uint64, limit, offset int64) ([]*AssetTransaction, error) {
	var result []*AssetTransaction
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (m *MockAssetTransactionModel) FindByUserIDAndType(ctx context.Context, userID uint64, transactionType int64, limit, offset int64) ([]*AssetTransaction, error) {
	var result []*AssetTransaction
	for _, tx := range m.transactions {
		if tx.UserID == userID && tx.Type == transactionType {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (m *MockAssetTransactionModel) CountByUserID(ctx context.Context, userID uint64) (int64, error) {
	count := int64(0)
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MockAssetTransactionModel) CountByUserIDAndType(ctx context.Context, userID uint64, transactionType int64) (int64, error) {
	count := int64(0)
	for _, tx := range m.transactions {
		if tx.UserID == userID && tx.Type == transactionType {
			count++
		}
	}
	return count, nil
}

func (m *MockAssetTransactionModel) FindOne(ctx context.Context, id uint64) (*AssetTransaction, error) {
	for _, tx := range m.transactions {
		if tx.ID == id {
			return tx, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockAssetTransactionModel) Update(ctx context.Context, data *AssetTransaction) error {
	if _, exists := m.transactions[data.TransactionID]; exists {
		m.transactions[data.TransactionID] = data
		return nil
	}
	return ErrNotFound
}

func (m *MockAssetTransactionModel) Delete(ctx context.Context, id uint64) error {
	for txID, tx := range m.transactions {
		if tx.ID == id {
			delete(m.transactions, txID)
			return nil
		}
	}
	return ErrNotFound
}

func TestAssetTransactionModel_Basic(t *testing.T) {
	model := NewMockAssetTransactionModel()
	ctx := context.Background()

	// 测试插入充值记录
	depositTx := &AssetTransaction{
		UserID:        1,
		TransactionID: "DEP_123456789",
		Currency:      "BTC",
		Type:          1, // 充值
		Amount:        "1.50000000",
		Fee:           "0.00000000",
		Status:        2, // 成功
		Address:       "",
		TxHash:        "",
		Remark:        "Deposit 1.5 BTC",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err := model.Insert(ctx, depositTx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), depositTx.ID)

	// 测试插入提现记录
	withdrawTx := &AssetTransaction{
		UserID:        1,
		TransactionID: "WTH_987654321",
		Currency:      "BTC",
		Type:          2, // 提现
		Amount:        "0.50000000",
		Fee:           "0.0005",
		Status:        1, // 待审核
		Address:       "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
		TxHash:        "",
		Remark:        "Withdraw 0.5 BTC to 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = model.Insert(ctx, withdrawTx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), withdrawTx.ID)

	// 测试根据交易ID查找
	foundTx, err := model.FindByTransactionID(ctx, "DEP_123456789")
	assert.NoError(t, err)
	assert.NotNil(t, foundTx)
	assert.Equal(t, "DEP_123456789", foundTx.TransactionID)
	assert.Equal(t, uint64(1), foundTx.UserID)
	assert.Equal(t, "BTC", foundTx.Currency)
	assert.Equal(t, int64(1), foundTx.Type)
	assert.Equal(t, "1.50000000", foundTx.Amount)

	// 测试根据用户ID查找所有交易
	userTxs, err := model.FindByUserID(ctx, 1, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, userTxs, 2)

	// 测试根据用户ID和类型查找交易
	depositTxs, err := model.FindByUserIDAndType(ctx, 1, 1, 10, 0) // 充值
	assert.NoError(t, err)
	assert.Len(t, depositTxs, 1)
	assert.Equal(t, "DEP_123456789", depositTxs[0].TransactionID)

	withdrawTxs, err := model.FindByUserIDAndType(ctx, 1, 2, 10, 0) // 提现
	assert.NoError(t, err)
	assert.Len(t, withdrawTxs, 1)
	assert.Equal(t, "WTH_987654321", withdrawTxs[0].TransactionID)

	// 测试统计用户交易数量
	totalCount, err := model.CountByUserID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), totalCount)

	depositCount, err := model.CountByUserIDAndType(ctx, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), depositCount)

	withdrawCount, err := model.CountByUserIDAndType(ctx, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), withdrawCount)

	// 测试查找不存在的交易
	_, err = model.FindByTransactionID(ctx, "NONEXISTENT")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestAssetTransactionModel_EdgeCases(t *testing.T) {
	model := NewMockAssetTransactionModel()
	ctx := context.Background()

	// 测试查找不存在用户的交易
	userTxs, err := model.FindByUserID(ctx, 999, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, userTxs, 0)

	// 测试统计不存在用户的交易数量
	count, err := model.CountByUserID(ctx, 999)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// 测试更新不存在的交易
	nonExistentTx := &AssetTransaction{
		ID:            999,
		TransactionID: "NONEXISTENT",
		UserID:        1,
		Currency:      "BTC",
		Type:          1,
		Amount:        "1.0",
		Fee:           "0.0",
		Status:        1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = model.Update(ctx, nonExistentTx)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	// 测试删除不存在的交易
	err = model.Delete(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestAssetTransactionModel_MultipleUsers(t *testing.T) {
	model := NewMockAssetTransactionModel()
	ctx := context.Background()

	// 为用户1创建交易
	user1Tx := &AssetTransaction{
		UserID:        1,
		TransactionID: "DEP_USER1",
		Currency:      "BTC",
		Type:          1,
		Amount:        "1.0",
		Fee:           "0.0",
		Status:        2,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 为用户2创建交易
	user2Tx := &AssetTransaction{
		UserID:        2,
		TransactionID: "DEP_USER2",
		Currency:      "ETH",
		Type:          1,
		Amount:        "10.0",
		Fee:           "0.0",
		Status:        2,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err := model.Insert(ctx, user1Tx)
	assert.NoError(t, err)

	_, err = model.Insert(ctx, user2Tx)
	assert.NoError(t, err)

	// 验证每个用户只能看到自己的交易
	user1Txs, err := model.FindByUserID(ctx, 1, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, user1Txs, 1)
	assert.Equal(t, "DEP_USER1", user1Txs[0].TransactionID)

	user2Txs, err := model.FindByUserID(ctx, 2, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, user2Txs, 1)
	assert.Equal(t, "DEP_USER2", user2Txs[0].TransactionID)

	// 验证统计数量正确
	user1Count, err := model.CountByUserID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user1Count)

	user2Count, err := model.CountByUserID(ctx, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user2Count)
}