package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/itkq/squadcast"
)

func main() {
	refreshToken := os.Getenv("SQUADCAST_REFRESH_TOKEN")
	if refreshToken == "" {
		log.Fatal("SQUADCAST_REFRESH_TOKEN is required")
	}

	client, err := squadcast.NewDefaultClient(refreshToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	services, err := client.GetAllServices(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, service := range services {
		fmt.Println(service.Name)
	}
}
