package handler

import (
	"errors"
	"net/http"

	"github.com/pomclaw/pomclaw/internal/logic"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Revoke skill from agent
func RevokeSkillHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("missing user ID"))
			return
		}

		skillID := r.PathValue("id")
		agentID := r.PathValue("agent_id")

		l := logic.NewRevokeSkillLogic(r.Context(), svcCtx)
		err := l.RevokeSkill(skillID, agentID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, map[string]string{"status": "revoked"})
		}
	}
}
