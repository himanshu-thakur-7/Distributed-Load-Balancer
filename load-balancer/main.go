/*
*
The load balancer will:

Accept incoming HTTP requests (/process)

Choose a backend node (round robin)

# Forward the request to that backend

# Return the backend’s response to the client

Log what happened
*
*/
package main

import (
	"io"
	"log"
	"net/http"
	"sync"
)

var backends = []string{
	"http://backend1:8080",
	"http://backend2:8080",
	"http://backend3:8080",
	"http://backend4:8080",
}

var (
	currentIndex int
	mu           sync.Mutex
)

func nextBackend() string {
	mu.Lock()
	defer mu.Unlock() // unlocks mutex in the end

	backend := backends[currentIndex]
	currentIndex = (currentIndex + 1) % len(backends)

	return backend
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	backend := nextBackend()
	targetURL := backend + "/process"

	log.Printf("[LB] forwarding request to %s\n", backend)

	// Create request to backend
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)

	if err != nil {
		http.Error(w, "failed to create backend request", http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		http.Error(w, "backend unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// copy backend response to client
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

}

func main() {
	http.HandleFunc("/process", handleRequest)

	log.Println("[LB] load balancer started on port 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
