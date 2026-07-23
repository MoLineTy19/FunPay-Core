package rest

import (
	"net/http"
)

// @Summary      Восстановить polling после auth_lost
// @Description  Посылает сигнал движку перечитать .env, обновить auth в памяти и реинициализировать runner. Работает только в состоянии auth_lost; в остальных случаях 409. Перед вызовом оператор должен обновить FP_GOLDEN_SEAL в .env.
// @Tags         control
// @Produce      json
// @Success      202  {object}  ControlResumeResponse
// @Failure      401  {object}  EngineError  "missing or invalid token"
// @Failure      409  {object}  EngineError  "conflict (не в состоянии auth_lost)"
// @Failure      503  {object}  EngineError  "service_unavailable (resume channel not configured)"
// @Security     ApiKeyAuth
// @Router       /control/resume [post]
func (s *Server) handleControlResume(w http.ResponseWriter, r *http.Request) {
	s.stateMu.RLock()
	state := s.state.Load().(string)
	s.stateMu.RUnlock()

	if state != "auth_lost" {
		writeEngineError(w, http.StatusConflict, "conflict",
			"resume is only valid when state=auth_lost (current: "+state+")", false)
		return
	}

	s.resumeMu.Lock()
	defer s.resumeMu.Unlock()
	if s.resumeCh == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable",
			"resume channel not configured", false)
		return
	}

	select {
	case s.resumeCh <- struct{}{}:
	default:
	}

	writeJSON(w, http.StatusAccepted, ControlResumeResponse{
		Status:  "accepted",
		Message: "resume signal sent; main will re-read .env and re-init runner",
	})
}
