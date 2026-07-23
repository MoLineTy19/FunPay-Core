package rest

import (
	"net/http"
	"time"
)

type AccountSnapshot struct {
	UserID   int64     `json:"userId"`
	Login    string    `json:"login"`
	Balance  string    `json:"balance"`
	LoadedAt time.Time `json:"loadedAt"`
}

// @Summary      Профиль продавца
// @Description  userId, login, баланс и время последней загрузки аккаунта.
// @Tags         account
// @Produce      json
// @Success      200  {object}  AccountSnapshot
// @Failure      401  {object}  EngineError  "missing or invalid token"
// @Failure      503  {object}  EngineError  "auth_lost"
// @Security     ApiKeyAuth
// @Router       /account [get]
func (s *Server) handleAccount(w http.ResponseWriter, r *http.Request) {
	snap := s.account.Load().(AccountSnapshot)
	writeJSON(w, http.StatusOK, snap)
}
