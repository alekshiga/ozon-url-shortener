package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/alekshiga/ozon-url-shortener/encoder"
	"github.com/alekshiga/ozon-url-shortener/handler"
	"github.com/alekshiga/ozon-url-shortener/service"
	"github.com/alekshiga/ozon-url-shortener/storage"
	"github.com/alekshiga/ozon-url-shortener/storage/memory"
	"github.com/alekshiga/ozon-url-shortener/storage/postgres"
)

func main() {
	storageType := flag.String(
		"storage",
		"memory",
		"storage type: memory or postgres",
	)

	flag.Parse()

	var store storage.Storage

	switch *storageType {
	case "memory":
		store = memory.New()

	case "postgres":
		connectionString := os.Getenv("DATABASE_URL")

		if connectionString == "" {
			log.Fatal("DATABASE_URL is not set")
		}

		postgresStore, err := postgres.New(
			context.Background(),
			connectionString,
		)
		if err != nil {
			log.Fatal(err)
		}

		defer postgresStore.Close()

		store = postgresStore

	default:
		log.Fatalf("unknown storage type: %s", *storageType)
	}

	coder := encoder.New()

	urlService := service.New(store, coder)
	urlHandler := handler.New(urlService)

	mux := http.NewServeMux()

	mux.HandleFunc("/shorten", urlHandler.Shorten)
	mux.HandleFunc("/urls/", urlHandler.GetOriginal)

	log.Printf(
		"server started on :8080, storage=%s",
		*storageType,
	)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
