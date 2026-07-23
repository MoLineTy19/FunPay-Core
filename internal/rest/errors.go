package rest

import (
	"encoding/json"
	"log"
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

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode: %v", err)
	}
}

func writeEngineError(w http.ResponseWriter, status int, code, msg string, retryable bool) {
	e := engineErrorJSON{}
	e.Error.Code = code
	e.Error.Message = msg
	e.Error.Retryable = retryable
	writeJSON(w, status, e)
}
