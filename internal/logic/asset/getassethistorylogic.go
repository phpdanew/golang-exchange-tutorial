package asset

import (
	"context"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssetHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAssetHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetHistoryLogic {
	return &GetAssetHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAssetHistoryLogic) GetAssetHistory(req *types.AssetHistoryRequest) (resp *types.AssetHistoryResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
