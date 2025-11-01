package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/coalyonysh/go-musthave-metrics/internal/handlers"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/gorilla/mux"
)

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Получаем значение по умолчанию из переменной окружения
	defaultAddr := getEnv("ADDRESS", "localhost:8080")

	// Определяем флаг с приоритетом переменной окружения как значения по умолчанию
	addr := flag.String("a", defaultAddr, "http server address")
	flag.Parse()

	// Финальный адрес сервера (флаг имеет высший приоритет)
	serverAddr := *addr

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

	log.Printf("Starting server on %s", serverAddr)
	log.Printf("Server available at http://%s", serverAddr)
	log.Printf("Available endpoints:")
	log.Printf("  POST /update/{type}/{name}/{value} - add metric")
	log.Printf("  GET  /value/{type}/{name} - get metric value")
	log.Printf("  GET  / - metrics dashboard")

	err := http.ListenAndServe(serverAddr, router)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
