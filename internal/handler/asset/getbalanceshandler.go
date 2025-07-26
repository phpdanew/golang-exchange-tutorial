package asset

import (
	"net/http"

	"crypto-exchange/internal/logic/asset"
	"crypto-exchange/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetBalancesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := asset.NewGetBalancesLogic(r.Context(), svcCtx)
		resp, err := l.GetBalances()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
