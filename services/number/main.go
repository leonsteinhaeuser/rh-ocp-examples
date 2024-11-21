package main

import (
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
)

var (
	envListenAddress = os.Getenv("LISTEN_ADDRESS")
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	if envListenAddress == "" {
		slog.Warn("LISTEN_ADDRESS is not set, using default value", "default", ":8081")
		envListenAddress = ":8081"
	}
}

func main() {
	// add routes
	http.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("GET /number", logMiddleware(getNumber))
	slog.Info("starting server on: " + envListenAddress)
	err := http.ListenAndServe(envListenAddress, nil)
	if err != nil {
		slog.Error("failed to start the server", "error", err)
		os.Exit(1)
	}
	slog.Warn("server stopped")
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
