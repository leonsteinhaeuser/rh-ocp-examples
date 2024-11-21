package main

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
)

var (
	envNumberServiceURL = os.Getenv("NUMBER_SERVICE_URL")
	envStatusServiceURL = os.Getenv("STATUS_SERVICE_URL")
	envListenAddress    = os.Getenv("LISTEN_ADDRESS")

	//go:embed view.html
	websiteHTML string
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	if envListenAddress == "" {
		slog.Warn("LISTEN_ADDRESS is not set, using default value", "default", ":8080")
		envListenAddress = ":8080"
	}
}

func main() {
	// add routes
	http.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("GET /api/v1/status", proxyEndpoint)
	http.HandleFunc("GET /", logMiddleware(getMain))
	slog.Info("starting server on port: " + envListenAddress)
	err := http.ListenAndServe(envListenAddress, nil)
	if err != nil {
		slog.Error("failed to start the server", "error", err)
		os.Exit(1)
	}
	slog.Warn("server stopped")
}

func proxyEndpoint(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, envStatusServiceURL, r.Body)
	if err != nil {
		slog.Error("could not create proxy request", "error", err)
		http.Error(w, "could not create proxy request", http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("could not forward request", "error", err)
		http.Error(w, "could not forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy headers from the response
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		slog.Error("could not copy response body", "error", err)
	}
}

var (
	websiteTemplate = template.Must(template.New("website").Parse(websiteHTML))
)

func logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.DebugContext(r.Context(), "incoming request", "method", r.Method, "url", r.URL.Path, "remote", r.RemoteAddr)
		next(w, r)
	}
}

func getMain(w http.ResponseWriter, r *http.Request) {
	rsp, err := http.Get(envNumberServiceURL + "/number")
	if err != nil || rsp.StatusCode != http.StatusOK {
		slog.DebugContext(r.Context(), "http upstream error: unable to get number from number service", "number_service_url", envNumberServiceURL+"/number", "error", err)
		http.Error(w, "http upstream error: could not get number from number service", http.StatusInternalServerError)
		return
	}
	defer rsp.Body.Close()

	payload := map[string]uint64{}
	err = json.NewDecoder(rsp.Body).Decode(&payload)
	if err != nil {
		slog.DebugContext(r.Context(), "http upstream error: could not decode number from number service", "error", err)
		http.Error(w, "http upstream error: could not decode number from number service", http.StatusInternalServerError)
		return
	}

	if _, ok := payload["number"]; !ok {
		slog.DebugContext(r.Context(), "http upstream error: number key not found in response from number service")
		http.Error(w, "http upstream error: number key not found in response from number service", http.StatusInternalServerError)
		return
	}

	err = websiteTemplate.Execute(w, payload)
	if err != nil {
		slog.DebugContext(r.Context(), "could not render website", "error", err)
		http.Error(w, "could not render website", http.StatusInternalServerError)
		return
	}
}
