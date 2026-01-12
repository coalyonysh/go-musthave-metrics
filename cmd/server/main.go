package main

import (
	"database/sql"
	"flag"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/handlers"
	"github.com/coalyonysh/go-musthave-metrics/internal/middleware"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool возвращает булево значение переменной окружения или значение по умолчанию
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvInt возвращает целочисленное значение переменной окружения или значение по умолчанию
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
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
var db *sql.DB

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

	// Определяем флаги со значениями по умолчанию
	addr := flag.String("a", "localhost:8080", "http server address")
	storeInterval := flag.Int("i", 300, "store interval in seconds (0 = sync)")
	filePath := flag.String("f", "/tmp/metrics-db.json", "file storage path")
	restore := flag.Bool("r", true, "restore metrics from file on startup")
	dsn := flag.String("d", "", "database DSN")
	flag.Parse()

	// Приоритет: переменная окружения > флаг > значение по умолчанию
	serverAddr := getEnv("ADDRESS", *addr)
	storeIntervalSec := getEnvInt("STORE_INTERVAL", *storeInterval)
	storagePath := getEnv("FILE_STORAGE_PATH", *filePath)
	shouldRestore := getEnvBool("RESTORE", *restore)
	databaseDSN := getEnv("DATABASE_DSN", *dsn)

	// Инициализация БД, если DSN указан
	if databaseDSN != "" {
		var err error
		db, err = sql.Open("sqlite3", databaseDSN)
		if err != nil {
			sugar.Fatalw("Failed to open database", "error", err, "dsn", databaseDSN)
		}
		if err = db.Ping(); err != nil {
			sugar.Fatalw("Failed to ping database", "error", err, "dsn", databaseDSN)
		}
		sugar.Infow("Database connected", "dsn", databaseDSN)
	}

	memStorage := storage.NewMemStorage()

	// Загружаем метрики при старте, если нужно
	if shouldRestore && storagePath != "" {
		if err := storage.LoadMetrics(memStorage, storagePath); err != nil {
			sugar.Warnw("Failed to load metrics from file", "error", err, "path", storagePath)
		} else {
			sugar.Infow("Metrics loaded from file", "path", storagePath)
		}
	}

	// Создаем функцию сохранения
	var saveMutex sync.Mutex
	saveFunc := func() {
		saveMutex.Lock()
		defer saveMutex.Unlock()
		if err := storage.SaveMetrics(memStorage, storagePath); err != nil {
			sugar.Errorw("Failed to save metrics", "error", err, "path", storagePath)
		} else {
			sugar.Debugw("Metrics saved", "path", storagePath)
		}
	}

	// Если interval=0, используем синхронное сохранение при каждом обновлении
	var finalStorage storage.Storage = memStorage
	if storeIntervalSec == 0 && storagePath != "" {
		finalStorage = storage.NewSyncStorage(memStorage, saveFunc)
		sugar.Infow("Synchronous saving enabled", "path", storagePath)
	} else if storeIntervalSec > 0 && storagePath != "" {
		// Запускаем периодическое сохранение
		go func() {
			ticker := time.NewTicker(time.Duration(storeIntervalSec) * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				saveFunc()
			}
		}()
		sugar.Infow("Periodic saving enabled", "interval", storeIntervalSec, "path", storagePath)
	}

	// Создаем хендлеры
	updateHandler := handlers.NewUpdateHandler(finalStorage)
	valueHandler := handlers.NewValueHandler(finalStorage)
	updateJSONHandler := handlers.NewUpdateJSONHandler(finalStorage)
	valueJSONHandler := handlers.NewValueJSONHandler(finalStorage)
	indexHandler := handlers.NewIndexHandler(finalStorage)
	pingHandler := handlers.NewPingHandler(db)

	// Создаем роутер с помощью gorilla/mux
	router := mux.NewRouter()

	// Добавляем middleware в правильном порядке:
	// 1. Сначала распаковка запросов (GzipDecompress)
	// 2. Затем сжатие ответов (GzipCompress)
	// 3. В конце логирование (WithLogging)
	router.Use(middleware.GzipDecompress)
	router.Use(middleware.GzipCompress)
	router.Use(WithLogging)

	// Регистрируем маршруты
	// Сначала регистрируем JSON эндпоинты (более специфичные по методу)
	// Используем Path() для точного сопоставления
	router.Path("/update").Handler(updateJSONHandler).Methods("POST")
	router.Path("/update/").Handler(updateJSONHandler).Methods("POST")
	router.Path("/value").Handler(valueJSONHandler).Methods("POST")
	router.Path("/value/").Handler(valueJSONHandler).Methods("POST")
	// Затем регистрируем маршруты с параметрами
	router.Handle("/update/{type}/{name}/{value}", updateHandler).Methods("POST")
	router.Handle("/value/{type}/{name}", valueHandler).Methods("GET")
	router.Handle("/ping", pingHandler).Methods("GET")
	router.Handle("/", indexHandler).Methods("GET")

	// записываем в лог, что сервер запускается
	sugar.Infow(
		"Starting server",
		"addr", serverAddr,
		"store_interval", storeIntervalSec,
		"file_path", storagePath,
		"restore", shouldRestore,
	)
	sugar.Infow(
		"Available endpoints",
		"POST /update/{type}/{name}/{value}", "add metric (URL params)",
		"POST /update", "add metric (JSON)",
		"GET /value/{type}/{name}", "get metric value (URL params)",
		"POST /value", "get metric value (JSON)",
		"GET /ping", "database ping",
		"GET /", "metrics dashboard",
	)

	// запускаем сервер
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		sugar.Fatalw(err.Error(), "event", "start server")
	}
}
