package admin

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTradingPairStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTradingPairStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradingPairStatsLogic {
	return &GetTradingPairStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTradingPairStatsLogic) GetTradingPairStats() (resp *types.TradingPairStatsResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
