package rest

import (
	"encoding/json"
	"net/http"
)

type engineErrorJSON struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		Retryable bool   `json:"retryable"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeEngineError(w http.ResponseWriter, status int, code, msg string, retryable bool) {
	e := engineErrorJSON{}
	e.Error.Code = code
	e.Error.Message = msg
	e.Error.Retryable = retryable
	writeJSON(w, status, e)
}
