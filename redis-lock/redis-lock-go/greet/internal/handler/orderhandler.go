package handler

import (
	"net/http"

	"github.com/tal-tech/go-zero/rest/httpx"
	"greet/internal/logic"
	"greet/internal/svc"
	"greet/internal/types"
)

func orderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewOrderLogic(r.Context(), svcCtx)
		resp, err := l.Order(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}