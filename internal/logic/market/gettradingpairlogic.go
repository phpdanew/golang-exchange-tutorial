package market

import (
	"context"
	"fmt"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTradingPairLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTradingPairLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradingPairLogic {
	return &GetTradingPairLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTradingPairLogic) GetTradingPair(symbol string) (resp *types.TradingPair, err error) {
	// 验证交易对符号格式
	if symbol == "" {
		return nil, fmt.Errorf("trading pair symbol is required")
	}

	// 从数据库查询交易对信息
	tradingPair, err := l.svcCtx.TradingPairModel.FindBySymbol(l.ctx, symbol)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("trading pair not found: %s", symbol)
		}
		l.Logger.Errorf("Failed to get trading pair %s: %v", symbol, err)
		return nil, err
	}

	// 转换为响应格式
	return &types.TradingPair{
		ID:            tradingPair.ID,
		Symbol:        tradingPair.Symbol,
		BaseCurrency:  tradingPair.BaseCurrency,
		QuoteCurrency: tradingPair.QuoteCurrency,
		MinAmount:     tradingPair.MinAmount,
		MaxAmount:     tradingPair.MaxAmount,
		PriceScale:    tradingPair.PriceScale,
		AmountScale:   tradingPair.AmountScale,
		Status:        tradingPair.Status,
		CreatedAt:     tradingPair.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
