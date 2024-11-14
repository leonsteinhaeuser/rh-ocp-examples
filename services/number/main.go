package main

import (
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"net/http"
)

func main() {
	// add routes
	http.HandleFunc("GET /number", getNumber)
	slog.Info("starting server on port 8081")
	http.ListenAndServe(":8081", nil)
}

func getNumber(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]uint64{"number": rand.Uint64()})
}
