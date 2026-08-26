package server

import (
	"encoding/json"
	"log"
	"net/http"

	"bannerfingerprint/internal/engine"
	"bannerfingerprint/internal/model"
)

// NewHandler returns the HTTP mux with the fingerprint endpoints wired up.
func NewHandler(e *engine.Engine) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/fingerprint", handleBatch(e))
	mux.HandleFunc("/fingerprint/single", handleSingle(e))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

func handleBatch(e *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var inputs []model.Input
		if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json: " + err.Error()})
			return
		}
		results := e.RecognizeBatch(inputs)
		writeJSON(w, http.StatusOK, results)
	}
}

func handleSingle(e *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var in model.Input
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json: " + err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, e.Recognize(in))
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}
