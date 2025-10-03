package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coalyonysh/go-musthave-metrics/internal/handlers"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
)

func main() {
	memStorage := storage.NewMemStorage()

	updateHandler := handlers.NewUpdateHandler(memStorage)

	mux := http.NewServeMux()
	mux.Handle("/update/", updateHandler)

	// Добавляем обработчик для корневого пути для проверки работы
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "Metrics Server is running!\n")
		fmt.Fprintf(w, "Use POST /update/<type>/<name>/<value> to submit metrics\n")
	})

	serverAddr := ":8080"
	log.Printf("Starting server on %s", serverAddr)
	log.Printf("Server available at http://localhost%s", serverAddr)

	err := http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
