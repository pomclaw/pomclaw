package handler

import (
	"errors"
	"net/http"

	"github.com/pomclaw/pomclaw/internal/logic"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// List LLM providers
func ListProvidersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("missing user ID"))
			return
		}

		l := logic.NewListProvidersLogic(r.Context(), svcCtx)
		resp, err := l.ListProviders(userID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
