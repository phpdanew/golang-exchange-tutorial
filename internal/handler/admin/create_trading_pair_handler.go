package admin

import (
	"net/http"

	"crypto-exchange/internal/logic/admin"
	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateTradingPairHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateTradingPairRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewCreateTradingPairLogic(r.Context(), svcCtx)
		resp, err := l.CreateTradingPair(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
