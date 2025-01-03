package producer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/go-redis/redis/v8"
)

// Diff represents changes in a file
type Diff struct {
	FilePath string `json:"file_path"`
	Change   string `json:"change"`
	Status   string `json:"status"`
}

// MergeRequest represents a GitLab merge request
type MergeRequest struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Diffs  []Diff `json:"diffs"`
	Status string `json:"status"`
}

var ctx = context.Background()

// AddMergeRequest dynamically adds a new merge request to the queue
func AddMergeRequest(rdb *redis.Client, mr MergeRequest) error {
	data, err := json.Marshal(mr)
	if err != nil {
		return err
	}

	queueKey := "merge_request_queue"
	if err := rdb.RPush(ctx, queueKey, data).Err(); err != nil {
		return err
	}

	log.Printf("Added Merge Request ID: %d, Title: %s to the queue\n", mr.ID, mr.Title)
	return nil
}

// InitializeProducer populates the queue with initial data
func InitializeProducer() error {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	// Example merge requests
	mergeRequests := []MergeRequest{
		{
			ID:     1,
			Title:  "Fix bug in authentication module",
			Diffs:  []Diff{{FilePath: "auth.go", Change: "Modified login function", Status: "Pending"}},
			Status: "Pending",
		},
		{
			ID:     2,
			Title:  "Improve performance of search",
			Diffs:  []Diff{{FilePath: "search.go", Change: "Optimized search algorithm", Status: "Pending"}},
			Status: "Pending",
		},
	}

	// Add each merge request to the queue
	for _, mr := range mergeRequests {
		if err := AddMergeRequest(rdb, mr); err != nil {
			return err
		}
	}
	return nil
}
