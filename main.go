package main

import (
	"log"

	"github.com/HenryArin/ReddmeitAlpha/services"
)

func main() {
	if err := services.RunInteractiveSession(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
