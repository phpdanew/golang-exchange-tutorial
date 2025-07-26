package market

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTradingPairsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTradingPairsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradingPairsLogic {
	return &GetTradingPairsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTradingPairsLogic) GetTradingPairs() (resp *types.TradingPairListResponse, err error) {
	// 获取所有活跃的交易对
	tradingPairs, err := l.svcCtx.TradingPairModel.FindActivePairs(l.ctx)
	if err != nil {
		l.Logger.Errorf("Failed to get active trading pairs: %v", err)
		return nil, err
	}

	// 转换为响应格式
	var pairs []types.TradingPair
	for _, pair := range tradingPairs {
		pairs = append(pairs, types.TradingPair{
			ID:            pair.ID,
			Symbol:        pair.Symbol,
			BaseCurrency:  pair.BaseCurrency,
			QuoteCurrency: pair.QuoteCurrency,
			MinAmount:     pair.MinAmount,
			MaxAmount:     pair.MaxAmount,
			PriceScale:    pair.PriceScale,
			AmountScale:   pair.AmountScale,
			Status:        pair.Status,
			CreatedAt:     pair.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.TradingPairListResponse{
		TradingPairs: pairs,
	}, nil
}
