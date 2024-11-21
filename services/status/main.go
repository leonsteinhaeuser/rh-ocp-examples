package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	// envExternalServicesToWatch is a comma-separated list of key=value services to watch (HTTP GET).
	// The key is the name of the service, and the value is the URL to check.
	envExternalServicesToWatch = os.Getenv("EXTERNAL_SERVICES_TO_WATCH")
	envCheckInterval           = os.Getenv("CHECK_INTERVAL")
	envRequestTimeout          = os.Getenv("REQUEST_TIMEOUT")
	envListenAddress           = os.Getenv("LISTEN_ADDRESS")

	externalServicesToWatch = map[string]string{}
	checkInterval           = 15 * time.Second
	requestTimeout          = 5 * time.Second

	mlock                  sync.Mutex = sync.Mutex{}
	externalServicesStatus            = map[string]ServiceStatus{}
)

type ServiceStatus struct {
	LastChecked time.Time `json:"last_checked"`
	Status      Status    `json:"status"`
	URL         string    `json:"url"`
}

type Status string

const (
	Unknown     Status = "unknown"
	Available   Status = "available"
	Unavailable Status = "unavailable"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	if envListenAddress == "" {
		slog.Warn("LISTEN_ADDRESS is not set, using default value", "default", ":8082")
		envListenAddress = ":8082"
	}

	// Parse the envExternalServicesToWatch and populate the externalServicesToWatch map.
	// If the envExternalServicesToWatch is empty, then we should watch the default services.
	if envExternalServicesToWatch == "" {
		externalServicesToWatch = map[string]string{
			"redhat": "https://www.redhat.com/en",
		}
	} else {
		services := strings.Split(envExternalServicesToWatch, ",")
		for _, service := range services {
			parts := strings.Split(service, "=")
			if len(parts) != 2 {
				// this is a fatal error, so we exit the program.
				slog.Error("invalid external service format, expected key=value format", "service", service)
				os.Exit(1)
			}
			externalServicesToWatch[parts[0]] = parts[1]
		}
	}

	// parse the envCheckInterval and set the checkInterval.
	if envCheckInterval != "" {
		d, err := time.ParseDuration(envCheckInterval)
		if err != nil {
			slog.Error("failed to parse the check interval", "interval", envCheckInterval, "error", err)
			os.Exit(1)
		}
		checkInterval = d
	}

	// parse the envRequestTimeout and set the requestTimeout.
	if envRequestTimeout != "" {
		d, err := time.ParseDuration(envRequestTimeout)
		if err != nil {
			slog.Error("failed to parse the request timeout", "timeout", envRequestTimeout, "error", err)
			os.Exit(1)
		}
		requestTimeout = d
	}
	http.DefaultClient.Timeout = requestTimeout
}

func main() {
	http.HandleFunc("GET /status", statusHandler)
	http.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	// check the availability of the external services.
	watcher(ctx)

	slog.Info("starting the server", "address", envListenAddress)
	err := http.ListenAndServe(envListenAddress, nil)
	if err != nil {
		slog.Error("failed to start the server", "error", err)
		os.Exit(1)
	}
	slog.Warn("server stopped")
}

func watcher(ctx context.Context) {
	for service, url := range externalServicesToWatch {
		go func(ctx context.Context, service, url string) {
			ticker := time.NewTicker(checkInterval)
			for {
				select {
				case <-ctx.Done():
					slog.Warn("received a signal to stop the watcher", "service", service)
					return
				case <-ticker.C:
					mlock.Lock()
					svc := ServiceStatus{
						LastChecked: time.Now(),
						Status:      Unknown,
						URL:         url,
					}
					err := checkExternalServiceAvailability(service, url)
					if err != nil {
						svc.Status = Unavailable
						externalServicesStatus[service] = svc
						mlock.Unlock()
						slog.Error("service is not available", "service", service, "error", err)
						continue
					}
					svc.Status = Available
					externalServicesStatus[service] = svc
					mlock.Unlock()
				}
			}
		}(ctx, service, url)
	}
}

// checkExternalServicesAvailability checks the availability of the external services.
// It sends an HTTP GET request to the URL of the service and checks the status code.
func checkExternalServiceAvailability(service string, url string) error {
	rsp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to check the availability of the service %s: %w", service, err)
	}
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("service %s is not available: %s", service, rsp.Status)
	}
	return nil
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(externalServicesStatus)
	if err != nil {
		slog.Error("failed to encode the response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
