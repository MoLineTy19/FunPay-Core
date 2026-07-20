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

func (s *Server) handleAccount(w http.ResponseWriter, r *http.Request) {
	snap := s.account.Load().(AccountSnapshot)
	writeJSON(w, http.StatusOK, snap)
}
