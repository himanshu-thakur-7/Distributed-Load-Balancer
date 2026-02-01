package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var isHealthy = true

type Response struct {
	NodeID           string `json:"node_id"`
	ProcessingTimeMs int64  `json:"processing_time_ms"`
}

func processHandler(nodeID string, rng *rand.Rand) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		sleepMs := rng.Intn(1500) + 500

		time.Sleep(time.Duration(sleepMs) * time.Millisecond)

		duration := time.Since(start).Milliseconds()

		log.Printf("[%s] processed request in %dms\n", nodeID, duration)

		response := Response{
			NodeID:           nodeID,
			ProcessingTimeMs: duration,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if isHealthy {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"unhealthy"}`))
	}
}

func toggleHealthHandler(w http.ResponseWriter, r *http.Request) {
	isHealthy = !isHealthy
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("health toggled"))
}

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "backend-unkown"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/process", processHandler(nodeID, rng))
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/toggle-health", toggleHealthHandler)

	log.Printf("[%s] starting backend server on port %s\n", nodeID, port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}

}
