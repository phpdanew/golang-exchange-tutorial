package trading

import (
	"net/http"

	"crypto-exchange/internal/logic/trading"
	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryOrdersRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := trading.NewQueryOrdersLogic(r.Context(), svcCtx)
		resp, err := l.QueryOrders(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
