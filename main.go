package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on env variables")
	}

	retentionStr := os.Getenv("RETENTION_DURATION")
	cleanupStr := os.Getenv("CLEANUP_INTERVAL")

	retention, err := time.ParseDuration(retentionStr)
	if err != nil {
		log.Fatalf("invalid RETENTION_DURATION: %v", err)
	}

	cleanupInterval, err := time.ParseDuration(cleanupStr)
	if err != nil {
		log.Fatalf("invalid CLEANUP_INTERVAL: %v", err)
	}

	cfg := &Config{
		ListenAddr: ":3000",
		StoreProducerFunc: func() Storer {
			return NewMemoryStore(retention)
		},
		CleanupInterval: cleanupInterval,
	}

	s, err := NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	s.Start()
}
