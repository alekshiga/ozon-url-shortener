package main

import (
	"log"
	"net/http"

	"github.com/alekshiga/ozon-url-shortener/encoder"
	"github.com/alekshiga/ozon-url-shortener/handler"
	"github.com/alekshiga/ozon-url-shortener/service"
	"github.com/alekshiga/ozon-url-shortener/storage/memory"
)

func main() {
	storage := memory.New()
	coder := encoder.New()

	urlService := service.New(storage, coder)
	urlHandler := handler.New(urlService)

	mux := http.NewServeMux()

	mux.HandleFunc("/shorten", urlHandler.Shorten)
	mux.HandleFunc("/urls/", urlHandler.GetOriginal)

	log.Println("server started on :8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
