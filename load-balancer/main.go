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
	"context"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/redis/go-redis/v9"
)

type Backend struct {
	ID          string
	URL         string
	ActiveConns int
}

var ctx = context.Background()

func newRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

}

func loadBackendsFromRedis(rdb *redis.Client) ([]Backend, error) {
	ids, err := rdb.SMembers(ctx, "backends").Result()

	if err != nil {
		return nil, err
	}

	var result []Backend

	for _, id := range ids {
		key := "backend:" + id

		data, err := rdb.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		if data["status"] != "healthy" {
			continue
		}

		result = append(result, Backend{
			ID:  id,
			URL: data["url"],
		})
	}

	return result, nil
}

var backends []Backend

var (
	currentIndex int
	mu           sync.Mutex
)

func nextBackend() Backend {
	mu.Lock()
	defer mu.Unlock() // unlocks mutex in the end

	backend := backends[currentIndex]
	currentIndex = (currentIndex + 1) % len(backends)

	return backend
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if len(backends) == 0 {
		http.Error(w, "no healthy backends available", http.StatusServiceUnavailable)
		return
	}

	backend := nextBackend()
	targetURL := backend.URL + "/process"

	log.Printf("[LB] forwarding request to %s URL: %s\n", backend.ID, backend.URL)

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
	rdb := newRedisClient()

	loadedBackends, err := loadBackendsFromRedis(rdb)

	if err != nil {
		log.Fatalf("failed to load backends: %v", err)
	}

	if len(loadedBackends) == 0 {
		log.Printf("no healthy backends found")
	}

	backends = loadedBackends
	log.Printf("[LB] loaded %d backends from redis", len(backends))

	http.HandleFunc("/process", handleRequest)

	log.Println("[LB] load balancer started on port 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
