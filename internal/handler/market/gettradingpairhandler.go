package market

import (
	"net/http"

	"crypto-exchange/internal/logic/market"
	"crypto-exchange/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

func GetTradingPairHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从URL路径中提取symbol参数
		vars := pathvar.Vars(r)
		symbol := vars["symbol"]
		
		l := market.NewGetTradingPairLogic(r.Context(), svcCtx)
		resp, err := l.GetTradingPair(symbol)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
