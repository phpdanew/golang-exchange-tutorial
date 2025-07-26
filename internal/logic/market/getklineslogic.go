package market

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKlinesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetKlinesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKlinesLogic {
	return &GetKlinesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKlinesLogic) GetKlines(req *types.KlineRequest) (resp *types.KlineResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
