package market

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllTickersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAllTickersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllTickersLogic {
	return &GetAllTickersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAllTickersLogic) GetAllTickers() (resp *types.AllTickersResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
