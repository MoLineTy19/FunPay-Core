package rest

import (
	"FunPay-Core/internal/engine"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// pollRequest — тело запроса long-poll событий.
type pollRequest struct {
	Since int64 `json:"since" example:"42"`
	Wait  int   `json:"wait" example:"15"`
}

type pollResponse struct {
	Events      []engine.Event `json:"events"`
	NextEventID int64          `json:"nextEventId,omitempty"`
}

// @Summary      Long-poll событий
// @Description  Возвращает события из буфера с eventId > since. Если событий нет, держит соединение до wait секунд (макс. 30). Если since указывает на вытесненное событие — 409 cursor_too_old (нужно пересобрать снимок через /orders, /account).
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        request  body      pollRequest  true  "Параметры long-poll"
// @Success      200      {object}  pollResponse
// @Failure      400      {object}  EngineError  "bad_request"
// @Failure      401      {object}  EngineError  "missing or invalid token"
// @Failure      409      {object}  EngineError  "cursor_too_old"
// @Failure      500      {object}  EngineError  "internal (retryable)"
// @Security     ApiKeyAuth
// @Router       /events/poll [post]
func (s *Server) handleEventsPoll(w http.ResponseWriter, r *http.Request) {
	var req pollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEngineError(w, 400, "bad_request", err.Error(), false)
		return
	}

	if req.Since < 0 {
		writeEngineError(w, 400, "bad_request", "since must be >= 0", false)
		return
	}

	if req.Wait < 0 {
		req.Wait = 0
	}
	if req.Wait > 30 {
		req.Wait = 30
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(req.Wait)*time.Second)
	defer cancel()

	ch := s.buf.Subscribe()
	defer s.buf.Unsubscribe(ch)

	for {
		events, err := s.buf.Since(req.Since)
		if errors.Is(err, engine.ErrCursorTooOld) {
			writeEngineError(w, 409, "cursor_too_old", err.Error(), false)
			return
		}
		if err != nil {
			writeEngineError(w, 500, "internal", err.Error(), true)
			return
		}
		if len(events) > 0 {
			next := events[len(events)-1].EventID
			writeJSON(w, 200, pollResponse{
				Events:      events,
				NextEventID: next,
			})
			return
		}

		select {
		case <-ch:
			continue
		case <-ctx.Done():
			writeJSON(w, 200, pollResponse{
				Events: []engine.Event{},
			})
			return
		}
	}
}
