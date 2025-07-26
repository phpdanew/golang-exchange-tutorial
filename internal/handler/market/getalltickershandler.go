package market

import (
	"net/http"

	"crypto-exchange/internal/logic/market"
	"crypto-exchange/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAllTickersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := market.NewGetAllTickersLogic(r.Context(), svcCtx)
		resp, err := l.GetAllTickers()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
