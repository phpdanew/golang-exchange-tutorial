package admin

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTradingPairLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTradingPairLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTradingPairLogic {
	return &CreateTradingPairLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTradingPairLogic) CreateTradingPair(req *types.CreateTradingPairRequest) (resp *types.TradingPair, err error) {
	// todo: add your logic here and delete this line

	return
}
