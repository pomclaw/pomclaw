package handler

import (
	"errors"
	"net/http"

	"github.com/pomclaw/pomclaw/internal/logic"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Update provider
func UpdateProviderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("missing user ID"))
			return
		}

		id := r.PathValue("id")

		var req types.UpdateProviderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUpdateProviderLogic(r.Context(), svcCtx)
		err := l.UpdateProvider(userID, id, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, map[string]string{"status": "updated"})
		}
	}
}
