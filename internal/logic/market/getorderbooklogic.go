package market

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderBookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderBookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderBookLogic {
	return &GetOrderBookLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderBookLogic) GetOrderBook(req *types.OrderBookRequest) (resp *types.OrderBookResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
