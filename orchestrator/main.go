package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func newRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
}

func getBackendIDs(rdb *redis.Client) ([]string, error) {
	return rdb.SMembers(ctx, "backends").Result()
}

func getBackendURL(rdb *redis.Client, backendID string) (string, error) {
	key := "backend:" + backendID
	return rdb.HGet(ctx, key, "url").Result()

}

func isBackendHealthy(url string) bool {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url + "/health")

	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func updateBackendStatus(rdb *redis.Client, backendId string, status string) {
	key := "backend:" + backendId

	_, err := rdb.HSet(ctx, key, map[string]interface{}{
		"status":       status,
		"last_checked": time.Now().Unix(),
	}).Result()

	if err != nil {
		log.Printf("[ORCH] failed to update %s: %v", backendId, err)
	}

}

func runHealthCheckCycle(rdb *redis.Client) {
	backendIDs, err := getBackendIDs(rdb)

	if err != nil {
		log.Printf("[ORCH] failed to get backends %v", err)
		return
	}

	for _, id := range backendIDs {
		url, err := getBackendURL(rdb, id)
		if err != nil {
			log.Printf("[ORCH] failed to get url for %s", id)
			continue
		}
		healthy := isBackendHealthy(url)

		status := "unhealthy"
		if healthy {
			status = "healthy"
		}

		updateBackendStatus(rdb, id, status)

		log.Printf("[ORCH] backend=%s status=%s", id, status)
	}

}

func main() {
	rdb := newRedisClient()

	log.Println("[ORCH] orchestrator started")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		runHealthCheckCycle(rdb)
		<-ticker.C
	}
}
