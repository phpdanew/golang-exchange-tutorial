package admin

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTradingPairLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTradingPairLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTradingPairLogic {
	return &UpdateTradingPairLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTradingPairLogic) UpdateTradingPair(req *types.UpdateTradingPairRequest) (resp *types.TradingPair, err error) {
	// todo: add your logic here and delete this line

	return
}
