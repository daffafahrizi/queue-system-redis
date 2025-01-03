package main

import (
	"log"
	"net/http"

	"redistest1/api"
	"redistest1/consumer"
	"redistest1/producer"

	"github.com/go-redis/redis/v8"
)

func main() {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Initialize producer with some initial data
	if err := producer.InitializeProducer(); err != nil {
		log.Fatalf("Error initializing producer: %v", err)
	}

	// Start consumer in a separate goroutine
	go func() {
		if err := consumer.ProcessMergeRequests(); err != nil {
			log.Fatalf("Error processing merge requests: %v", err)
		}
	}()

	// Define the HTTP handlers
	http.HandleFunc("/merge-requests", api.AddMultipleMergeRequestsHandler(rdb))
	http.HandleFunc("/merge-requests/status", api.GetMergeRequestsStatusHandler(rdb))

	// Start the HTTP server
	log.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
