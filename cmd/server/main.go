package main

import (
	"fmt"
	"log"
	"net/http"
	"time-tracker/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}
	fmt.Println(":" + cfg.Port)

	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
