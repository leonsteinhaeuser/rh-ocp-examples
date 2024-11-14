package main

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
)

var (
	envNumberServiceURL = os.Getenv("NUMBER_SERVICE_URL")
)

func main() {
	// add routes
	http.HandleFunc("GET /", getMain)
	slog.Info("starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}

const (
	websiteHTML = `<!DOCTYPE html>
<html>
	<head>
		<title>View Service</title>
	</head>
	<body>
		<h1>Hello, World!</h1>
		<p>This is the view service.</p>
		<h5>The number you requested is:</h5><p>{{.number}}</p>
	</body>
</html>`
)

var (
	websiteTemplate = template.Must(template.New("website").Parse(websiteHTML))
)

func getMain(w http.ResponseWriter, r *http.Request) {
	rsp, err := http.Get(envNumberServiceURL + "/number")
	if err != nil || rsp.StatusCode != http.StatusOK {
		http.Error(w, "http upstream error: could not get number from number service", http.StatusInternalServerError)
		return
	}
	defer rsp.Body.Close()

	payload := map[string]uint64{}
	err = json.NewDecoder(rsp.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "http upstream error: could not decode number from number service", http.StatusInternalServerError)
		return
	}

	if _, ok := payload["number"]; !ok {
		http.Error(w, "http upstream error: number key not found in response from number service", http.StatusInternalServerError)
		return
	}

	err = websiteTemplate.Execute(w, payload)
	if err != nil {
		http.Error(w, "could not render website", http.StatusInternalServerError)
		return
	}
}
