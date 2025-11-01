package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/handlers"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// responseData для хранения сведений об ответе
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter добавляет реализацию http.ResponseWriter
type loggingResponseWriter struct {
	http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
	responseData        *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// WithLogging добавляет дополнительный код для регистрации сведений о запросе
func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		// функция Now() возвращает текущее время
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		// точка, где выполняется хендлер
		h.ServeHTTP(&lw, r) // обслуживание оригинального запроса

		// Since возвращает разницу во времени между start и моментом вызова Since
		duration := time.Since(start)

		// отправляем сведения о запросе в zap
		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)
	}
	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}

var sugar zap.SugaredLogger

func main() {
	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	sugar = *logger.Sugar()

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

	// Добавляем middleware логирования ко всем маршрутам
	router.Use(WithLogging)

	// Регистрируем маршруты
	router.Handle("/update/{type}/{name}/{value}", updateHandler).Methods("POST")
	router.Handle("/value/{type}/{name}", valueHandler).Methods("GET")
	router.Handle("/", indexHandler).Methods("GET")

	// записываем в лог, что сервер запускается
	sugar.Infow(
		"Starting server",
		"addr", serverAddr,
	)
	sugar.Infow(
		"Available endpoints",
		"POST /update/{type}/{name}/{value}", "add metric",
		"GET /value/{type}/{name}", "get metric value",
		"GET /", "metrics dashboard",
	)

	// запускаем сервер
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		sugar.Fatalw(err.Error(), "event", "start server")
	}
}
