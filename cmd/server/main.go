package main

import (
	"log"
	"net/http"

	"github.com/coalyonysh/go-musthave-metrics/internal/handlers"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/gorilla/mux"
)

func main() {
	memStorage := storage.NewMemStorage()

	// Создаем хендлеры
	updateHandler := handlers.NewUpdateHandler(memStorage)
	valueHandler := handlers.NewValueHandler(memStorage)
	indexHandler := handlers.NewIndexHandler(memStorage)

	// Создаем роутер с помощью gorilla/mux
	router := mux.NewRouter()

	// Регистрируем маршруты
	router.Handle("/update/{type}/{name}/{value}", updateHandler).Methods("POST")
	router.Handle("/value/{type}/{name}", valueHandler).Methods("GET")
	router.Handle("/", indexHandler).Methods("GET")

	serverAddr := ":8080"
	log.Printf("Starting server on %s", serverAddr)
	log.Printf("Server available at http://localhost%s", serverAddr)
	log.Printf("Available endpoints:")
	log.Printf("  POST /update/{type}/{name}/{value} - add metric")
	log.Printf("  GET  /value/{type}/{name} - get metric value")
	log.Printf("  GET  / - metrics dashboard")

	err := http.ListenAndServe(serverAddr, router)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
