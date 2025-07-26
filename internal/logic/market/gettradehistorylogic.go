package market

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTradeHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTradeHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradeHistoryLogic {
	return &GetTradeHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTradeHistoryLogic) GetTradeHistory(req *types.TradeHistoryRequest) (resp *types.TradeHistoryResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
