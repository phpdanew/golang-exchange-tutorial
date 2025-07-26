package market

import (
	"net/http"

	"crypto-exchange/internal/logic/market"
	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetTradeHistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TradeHistoryRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := market.NewGetTradeHistoryLogic(r.Context(), svcCtx)
		resp, err := l.GetTradeHistory(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
