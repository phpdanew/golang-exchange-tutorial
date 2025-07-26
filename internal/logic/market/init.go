package market

import (
	"context"

	"crypto-exchange/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// InitializeTradingPairs 初始化交易对数据
// 这个函数应该在应用启动时调用，用于设置默认的交易对配置
func InitializeTradingPairs(ctx context.Context, svcCtx *svc.ServiceContext) error {
	manager := NewTradingPairManager(ctx, svcCtx)
	
	logx.Info("Initializing default trading pairs...")
	
	err := manager.InitializeDefaultTradingPairs()
	if err != nil {
		logx.Errorf("Failed to initialize trading pairs: %v", err)
		return err
	}
	
	logx.Info("Trading pairs initialization completed successfully")
	return nil
}

// GetTradingPairStats 获取交易对统计信息的便捷函数
func GetTradingPairStats(ctx context.Context, svcCtx *svc.ServiceContext) (map[string]interface{}, error) {
	manager := NewTradingPairManager(ctx, svcCtx)
	return manager.GetTradingPairStats()
}