package trading

import (
	"net/http"

	"crypto-exchange/internal/logic/trading"
	"crypto-exchange/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := trading.NewGetOrderLogic(r.Context(), svcCtx)
		
		// 从路径参数中获取订单ID
		orderID := r.URL.Query().Get("id")
		if orderID == "" {
			// 尝试从路径参数中获取
			orderID = r.PathValue("id")
		}
		
		resp, err := l.GetOrder(orderID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
