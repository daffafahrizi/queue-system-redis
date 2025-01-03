package api

import (
	"encoding/json"
	"net/http"
	"fmt"
	"context"

	"redistest1/consumer"
	"redistest1/producer"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// AddMultipleMergeRequestsHandler handles the POST request to add multiple merge requests
func AddMultipleMergeRequestsHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var newMRs []producer.MergeRequest

		// Parse the JSON body into a slice of MergeRequest
		if err := json.NewDecoder(r.Body).Decode(&newMRs); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Add each merge request to the Redis queue
		for _, mr := range newMRs {
			if err := producer.AddMergeRequest(rdb, mr); err != nil {
				http.Error(w, "Failed to add one or more merge requests", http.StatusInternalServerError)
				return
			}
		}

		// Trigger the consumer to process the queue immediately
		go func() {
			if err := consumer.ProcessMergeRequests(); err != nil {
				// Log the error if processing fails
				// In production, consider adding retry logic
			}
		}()

		// Respond with success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":         "Merge requests added successfully and processing started",
			"merge_requests":  newMRs,
		})
	}
	
}

// GetMergeRequestsStatusHandler handles the GET request to list all merge requests and their statuses
func GetMergeRequestsStatusHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var mergeRequests []producer.MergeRequest

		// 1. Fetch merge requests from Redis queue (waiting to be processed)
		queueKey := "merge_request_queue"
		queueItems, err := rdb.LRange(ctx, queueKey, 0, -1).Result()
		if err != nil {
			http.Error(w, "Failed to fetch merge requests from queue", http.StatusInternalServerError)
			return
		}

		for _, item := range queueItems {
			var mr producer.MergeRequest
			if err := json.Unmarshal([]byte(item), &mr); err != nil {
				http.Error(w, "Failed to parse merge request data from queue", http.StatusInternalServerError)
				return
			}
			mergeRequests = append(mergeRequests, mr)
		}

		// 2. Fetch merge requests that are being processed or completed
		keys, err := rdb.Keys(ctx, "merge_request:*").Result()
		if err != nil {
			http.Error(w, "Failed to fetch merge request keys", http.StatusInternalServerError)
			return
		}

		for _, key := range keys {
			data, err := rdb.HGet(ctx, key, "data").Result()
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to fetch merge request data for key: %s", key), http.StatusInternalServerError)
				return
			}

			var mr producer.MergeRequest
			if err := json.Unmarshal([]byte(data), &mr); err != nil {
				http.Error(w, "Failed to parse merge request data", http.StatusInternalServerError)
				return
			}

			mergeRequests = append(mergeRequests, mr)
		}

		// Respond with the list of merge requests and their statuses
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"merge_requests": mergeRequests,
		})
	}
}

