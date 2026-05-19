package handler

import (
	"errors"
	"net/http"

	"github.com/pomclaw/pomclaw/internal/logic"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// List agent skills
func ListAgentSkillsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("missing user ID"))
			return
		}

		agentID := r.PathValue("agent_id")

		l := logic.NewListAgentSkillsLogic(r.Context(), svcCtx)
		resp, err := l.ListAgentSkills(userID, agentID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
