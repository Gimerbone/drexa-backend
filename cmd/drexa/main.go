package main

import (
	"context"
	"drexa/internal/config"
	"drexa/internal/infrastructure/database"
	"fmt"
	"log"
	"os"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Println(err)
	}

	srv := NewServer(config.Load(), db)

	ctx := context.Background()
	if err := srv.Start(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
