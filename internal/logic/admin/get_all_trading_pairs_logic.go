package admin

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllTradingPairsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAllTradingPairsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllTradingPairsLogic {
	return &GetAllTradingPairsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAllTradingPairsLogic) GetAllTradingPairs() (resp *types.TradingPairListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
