package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// Diff represents changes in a file
type Diff struct {
	FilePath string `json:"file_path"`
	Change   string `json:"change"`
	Status   string `json:"status"` // Status of the diff
}

// MergeRequest represents a GitLab merge request
type MergeRequest struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Diffs  []Diff `json:"diffs"`
	Status string `json:"status"` // Status of the merge request
}

var ctx = context.Background()

// ProcessMergeRequests processes and dequeues merge requests from the Redis queue
func ProcessMergeRequests() error {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	queueKey := "merge_request_queue"
	for {
		data, err := rdb.LPop(ctx, queueKey).Result()
		if err == redis.Nil {
			log.Println("Queue is empty.")
			break
		} else if err != nil {
			return err
		}

		var mr MergeRequest
		if err := json.Unmarshal([]byte(data), &mr); err != nil {
			return err
		}

		// Store the merge request in Redis for status tracking
		mrKey := fmt.Sprintf("merge_request:%d", mr.ID)
		if err := updateMergeRequestStatus(rdb, mrKey, mr); err != nil {
			return err
		}

		fmt.Printf("Processing Merge Request ID: %d, Title: %s, Status: %s\n", mr.ID, mr.Title, mr.Status)

		for i, diff := range mr.Diffs {
			fmt.Printf("\tFile: %s, Change: %s, Status: %s\n", diff.FilePath, diff.Change, diff.Status)

			// Simulate processing the diff with a delay
			fmt.Printf("\tAI is processing diff in file: %s\n", diff.FilePath)
			time.Sleep(15 * time.Second) // Simulate processing delay

			// Update the diff status
			mr.Diffs[i].Status = "Completed"

			// Update the status in Redis
			if err := updateMergeRequestStatus(rdb, mrKey, mr); err != nil {
				return err
			}

			fmt.Printf("\tUpdated Status: %s\n", mr.Diffs[i].Status)
		}

		// Update merge request status after processing all diffs
		mr.Status = "Completed"
		if err := updateMergeRequestStatus(rdb, mrKey, mr); err != nil {
			return err
		}

		fmt.Printf("Merge Request %d Status Updated to: %s\n", mr.ID, mr.Status)
	}
	return nil
}


// updateMergeRequestStatus updates the merge request status in Redis
func updateMergeRequestStatus(rdb *redis.Client, mrKey string, mr MergeRequest) error {
	data, err := json.Marshal(mr)
	if err != nil {
		return err
	}
	return rdb.HSet(ctx, mrKey, "data", data).Err()
}
