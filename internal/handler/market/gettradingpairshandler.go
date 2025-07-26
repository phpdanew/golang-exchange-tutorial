package market

import (
	"net/http"

	"crypto-exchange/internal/logic/market"
	"crypto-exchange/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetTradingPairsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := market.NewGetTradingPairsLogic(r.Context(), svcCtx)
		resp, err := l.GetTradingPairs()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
