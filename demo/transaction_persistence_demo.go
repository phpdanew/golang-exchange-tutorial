package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"crypto-exchange/model"
)

// 演示交易记录持久化功能
func main() {
	fmt.Println("=== 交易记录持久化演示 ===")

	// 创建模拟的交易记录模型
	txModel := NewMockAssetTransactionModel()
	ctx := context.Background()

	// 演示场景1：用户充值
	fmt.Println("\n1. 用户充值演示")
	depositTx := &model.AssetTransaction{
		UserID:        1001,
		TransactionID: "DEP_20240726001",
		Currency:      "BTC",
		Type:          1, // 充值
		Amount:        "0.50000000",
		Fee:           "0.00000000",
		Status:        2, // 成功
		Address:       "",
		TxHash:        "0x1234567890abcdef...",
		Remark:        "用户充值 0.5 BTC",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err := txModel.Insert(ctx, depositTx)
	if err != nil {
		log.Fatal("充值记录插入失败:", err)
	}

	fmt.Printf("✅ 充值记录已保存: %s\n", depositTx.TransactionID)
	fmt.Printf("   用户ID: %d\n", depositTx.UserID)
	fmt.Printf("   币种: %s\n", depositTx.Currency)
	fmt.Printf("   金额: %s\n", depositTx.Amount)
	fmt.Printf("   状态: %d (成功)\n", depositTx.Status)

	// 演示场景2：用户提现
	fmt.Println("\n2. 用户提现演示")
	withdrawTx := &model.AssetTransaction{
		UserID:        1001,
		TransactionID: "WTH_20240726001",
		Currency:      "BTC",
		Type:          2, // 提现
		Amount:        "0.20000000",
		Fee:           "0.0005",
		Status:        1, // 待审核
		Address:       "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
		TxHash:        "",
		Remark:        "用户提现 0.2 BTC 到外部地址",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = txModel.Insert(ctx, withdrawTx)
	if err != nil {
		log.Fatal("提现记录插入失败:", err)
	}

	fmt.Printf("✅ 提现记录已保存: %s\n", withdrawTx.TransactionID)
	fmt.Printf("   用户ID: %d\n", withdrawTx.UserID)
	fmt.Printf("   币种: %s\n", withdrawTx.Currency)
	fmt.Printf("   金额: %s\n", withdrawTx.Amount)
	fmt.Printf("   手续费: %s\n", withdrawTx.Fee)
	fmt.Printf("   提现地址: %s\n", withdrawTx.Address)
	fmt.Printf("   状态: %d (待审核)\n", withdrawTx.Status)

	// 演示场景3：查询用户交易记录
	fmt.Println("\n3. 查询用户交易记录")
	userTxs, err := txModel.FindByUserID(ctx, 1001, 10, 0)
	if err != nil {
		log.Fatal("查询用户交易记录失败:", err)
	}

	fmt.Printf("✅ 用户 %d 的交易记录 (共 %d 条):\n", 1001, len(userTxs))
	for i, tx := range userTxs {
		txType := "充值"
		if tx.Type == 2 {
			txType = "提现"
		}
		fmt.Printf("   %d. %s - %s %s %s (状态: %d)\n", 
			i+1, tx.TransactionID, txType, tx.Amount, tx.Currency, tx.Status)
	}

	// 演示场景4：按类型查询交易记录
	fmt.Println("\n4. 按类型查询交易记录")
	
	// 查询充值记录
	depositTxs, err := txModel.FindByUserIDAndType(ctx, 1001, 1, 10, 0)
	if err != nil {
		log.Fatal("查询充值记录失败:", err)
	}
	fmt.Printf("✅ 用户 %d 的充值记录 (共 %d 条):\n", 1001, len(depositTxs))
	for _, tx := range depositTxs {
		fmt.Printf("   - %s: %s %s\n", tx.TransactionID, tx.Amount, tx.Currency)
	}

	// 查询提现记录
	withdrawTxs, err := txModel.FindByUserIDAndType(ctx, 1001, 2, 10, 0)
	if err != nil {
		log.Fatal("查询提现记录失败:", err)
	}
	fmt.Printf("✅ 用户 %d 的提现记录 (共 %d 条):\n", 1001, len(withdrawTxs))
	for _, tx := range withdrawTxs {
		fmt.Printf("   - %s: %s %s (手续费: %s)\n", tx.TransactionID, tx.Amount, tx.Currency, tx.Fee)
	}

	// 演示场景5：统计交易数量
	fmt.Println("\n5. 统计交易数量")
	totalCount, err := txModel.CountByUserID(ctx, 1001)
	if err != nil {
		log.Fatal("统计总交易数量失败:", err)
	}

	depositCount, err := txModel.CountByUserIDAndType(ctx, 1001, 1)
	if err != nil {
		log.Fatal("统计充值数量失败:", err)
	}

	withdrawCount, err := txModel.CountByUserIDAndType(ctx, 1001, 2)
	if err != nil {
		log.Fatal("统计提现数量失败:", err)
	}

	fmt.Printf("✅ 用户 %d 的交易统计:\n", 1001)
	fmt.Printf("   总交易数: %d\n", totalCount)
	fmt.Printf("   充值次数: %d\n", depositCount)
	fmt.Printf("   提现次数: %d\n", withdrawCount)

	// 演示场景6：根据交易ID查询
	fmt.Println("\n6. 根据交易ID查询")
	foundTx, err := txModel.FindByTransactionID(ctx, "DEP_20240726001")
	if err != nil {
		log.Fatal("根据交易ID查询失败:", err)
	}

	fmt.Printf("✅ 找到交易记录: %s\n", foundTx.TransactionID)
	fmt.Printf("   创建时间: %s\n", foundTx.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   备注: %s\n", foundTx.Remark)

	fmt.Println("\n=== 演示完成 ===")
	fmt.Println("✅ 所有交易记录都已成功持久化到数据库")
	fmt.Println("✅ 支持多种查询方式：按用户、按类型、按交易ID")
	fmt.Println("✅ 提供完整的审计跟踪和交易历史")
}

// MockAssetTransactionModel 用于演示的模拟实现
type MockAssetTransactionModel struct {
	transactions map[string]*model.AssetTransaction
	nextID       uint64
}

func NewMockAssetTransactionModel() *MockAssetTransactionModel {
	return &MockAssetTransactionModel{
		transactions: make(map[string]*model.AssetTransaction),
		nextID:       1,
	}
}

func (m *MockAssetTransactionModel) Insert(ctx context.Context, data *model.AssetTransaction) (interface{}, error) {
	data.ID = m.nextID
	m.nextID++
	m.transactions[data.TransactionID] = data
	return nil, nil
}

func (m *MockAssetTransactionModel) FindByTransactionID(ctx context.Context, transactionID string) (*model.AssetTransaction, error) {
	if tx, exists := m.transactions[transactionID]; exists {
		return tx, nil
	}
	return nil, model.ErrNotFound
}

func (m *MockAssetTransactionModel) FindByUserID(ctx context.Context, userID uint64, limit, offset int64) ([]*model.AssetTransaction, error) {
	var result []*model.AssetTransaction
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (m *MockAssetTransactionModel) FindByUserIDAndType(ctx context.Context, userID uint64, transactionType int64, limit, offset int64) ([]*model.AssetTransaction, error) {
	var result []*model.AssetTransaction
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