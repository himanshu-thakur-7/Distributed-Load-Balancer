package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

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

	log.Printf("[%s] starting backend server on port %s\n", nodeID, port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}

}
