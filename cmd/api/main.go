package main

import (
	"log"

	"go-backend-project/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
