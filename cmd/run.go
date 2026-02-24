package cmd

import (
	"log"
	"own-redis/internal/server"
	"own-redis/internal/store"
)

func Run() {
	port := parseFlags()

	store := store.NewStore()

	if err := server.Start(port, store); err != nil {
		log.Fatal(err)
	}
}
