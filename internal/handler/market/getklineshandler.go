package market

import (
	"net/http"

	"crypto-exchange/internal/logic/market"
	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetKlinesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.KlineRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := market.NewGetKlinesLogic(r.Context(), svcCtx)
		resp, err := l.GetKlines(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
