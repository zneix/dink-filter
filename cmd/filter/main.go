package main

import (
	"log"

	"github.com/zneix/dink-filter/internal/api"
	"github.com/zneix/dink-filter/internal/config"
)

func init() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)
}

func main() {
	log.Println("Starting dink filter")

	cfg := config.New()

	api := api.New(cfg)
	api.Listen()
}
