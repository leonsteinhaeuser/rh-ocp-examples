package main

import (
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)
}

func main() {
	// add routes
	http.HandleFunc("GET /number", logMiddleware(getNumber))
	slog.Info("starting server on port 8081")
	http.ListenAndServe(":8081", nil)
}

func logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.DebugContext(r.Context(), "incoming request", "method", r.Method, "url", r.URL.Path, "remote", r.RemoteAddr)
		next(w, r)
	}
}

func getNumber(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]uint64{"number": rand.Uint64()})
}
