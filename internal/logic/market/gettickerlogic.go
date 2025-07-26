package market

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTickerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTickerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTickerLogic {
	return &GetTickerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTickerLogic) GetTicker() (resp *types.TickerResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
