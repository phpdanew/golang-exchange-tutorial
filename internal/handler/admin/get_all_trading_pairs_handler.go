package admin

import (
	"net/http"

	"crypto-exchange/internal/logic/admin"
	"crypto-exchange/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAllTradingPairsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := admin.NewGetAllTradingPairsLogic(r.Context(), svcCtx)
		resp, err := l.GetAllTradingPairs()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
