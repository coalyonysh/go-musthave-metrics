package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/coalyonysh/go-musthave-metrics/internal/buildinfo"
	"github.com/coalyonysh/go-musthave-metrics/internal/handlers"
	"github.com/coalyonysh/go-musthave-metrics/internal/middleware"
	"github.com/coalyonysh/go-musthave-metrics/internal/service"
	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
	"github.com/coalyonysh/go-musthave-metrics/pkg/crypto"
	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Глобальные переменные для информации о сборке
var (
	buildVersion string
	buildDate    string
	buildCommit  string
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

// createDecryptMiddleware создает middleware для дешифрования запросов
func createDecryptMiddleware(privateKey *crypto.PrivateKey) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем заголовок X-Encrypted
			if r.Header.Get("X-Encrypted") == "true" {
				// Читаем тело запроса
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Failed to read request body", http.StatusBadRequest)
					return
				}

				// Дешифруем данные
				decrypted, err := privateKey.Decrypt(body)
				if err != nil {
					sugar.Warnw("Failed to decrypt request", "error", err)
					http.Error(w, "Failed to decrypt request", http.StatusBadRequest)
					return
				}

				// Заменяем тело запроса на дешифрованные данные
				r.Body = io.NopCloser(bytes.NewReader(decrypted))
				r.ContentLength = int64(len(decrypted))
				r.Header.Set("Content-Length", strconv.Itoa(len(decrypted)))
			}
			h.ServeHTTP(w, r)
		})
	}
}

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
	keyFile := flag.String("k", "", "path to file containing hash key")
	auditFile := flag.String("audit-file", "", "path to audit log file")
	auditURL := flag.String("audit-url", "", "URL to send audit logs")
	cryptoKeyFile := flag.String("crypto-key", "", "path to RSA private key file for decryption")
	flag.Parse()

	// Вывод информации о сборке
	buildinfo.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	// Приоритет: переменная окружения > флаг > значение по умолчанию
	serverAddr := getEnv("ADDRESS", *addr)
	storeIntervalSec := getEnvInt("STORE_INTERVAL", *storeInterval)
	storagePath := getEnv("FILE_STORAGE_PATH", *filePath)
	shouldRestore := getEnvBool("RESTORE", *restore)
	databaseDSN := getEnv("DATABASE_DSN", *dsn)
	keyFileEnv := getEnv("KEY", *keyFile)
	auditFilePath := getEnv("AUDIT_FILE", *auditFile)
	auditURLPath := getEnv("AUDIT_URL", *auditURL)
	cryptoKeyPath := getEnv("CRYPTO_KEY", *cryptoKeyFile)

	// Читаем приватный ключ для дешифрования, если указан
	var privateKey *crypto.PrivateKey
	if cryptoKeyPath != "" {
		var err error
		privateKey, err = crypto.LoadPrivateKey(cryptoKeyPath)
		if err != nil {
			sugar.Warnw("Failed to load private key", "error", err, "path", cryptoKeyPath)
			privateKey = nil
		} else {
			sugar.Infow("Private key loaded", "path", cryptoKeyPath)
		}
	}

	// Читаем ключ из файла, если указан
	var hashKey string
	if keyFileEnv != "" {
		keyBytes, err := os.ReadFile(keyFileEnv)
		if err != nil {
			// Если файл не найден, используем значение как ключ напрямую
			hashKey = keyFileEnv
		} else {
			hashKey = string(keyBytes)
		}
	}

	// Создаем сервис аудита
	auditService := service.NewAuditService()
	if auditFilePath != "" {
		fileObserver := service.NewFileObserver(auditFilePath)
		auditService.AddObserver(fileObserver)
		sugar.Infow("Audit file observer added", "path", auditFilePath)
	}
	if auditURLPath != "" {
		urlObserver := service.NewURLOobserver(auditURLPath)
		auditService.AddObserver(urlObserver)
		sugar.Infow("Audit URL observer added", "url", auditURLPath)
	}

	// Инициализация БД, если DSN указан
	if databaseDSN != "" {
		var err error
		db, err = sql.Open("postgres", databaseDSN)
		if err != nil {
			sugar.Fatalw("Failed to open database", "error", err, "dsn", databaseDSN)
		}
		if err = db.Ping(); err != nil {
			sugar.Fatalw("Failed to ping database", "error", err, "dsn", databaseDSN)
		}
		sugar.Infow("Database connected", "dsn", databaseDSN)

		// Запуск миграций
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			sugar.Fatalw("Failed to create migration driver", "error", err)
		}
		m, err := migrate.NewWithDatabaseInstance(
			"file://migrations",
			"postgres", driver)
		if err != nil {
			sugar.Fatalw("Failed to create migrate instance", "error", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			sugar.Fatalw("Failed to run migrations", "error", err)
		}
		sugar.Infow("Migrations completed")
	}

	var finalStorage storage.Storage

	if databaseDSN != "" {
		finalStorage = storage.NewDBStorage(db)
		sugar.Infow("Using database storage")
	} else {
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
		finalStorage = memStorage
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
	}

	// Создаем хендлеры
	updateHandler := handlers.NewUpdateHandler(finalStorage, auditService)
	valueHandler := handlers.NewValueHandler(finalStorage)
	updateJSONHandler := handlers.NewUpdateJSONHandler(finalStorage, hashKey, auditService)
	valueJSONHandler := handlers.NewValueJSONHandler(finalStorage, hashKey)
	indexHandler := handlers.NewIndexHandler(finalStorage)
	pingHandler := handlers.NewPingHandler(db)
	updatesHandler := handlers.NewUpdatesHandler(finalStorage, hashKey, auditService)

	// Middleware для дешифрования запросов от агента
	var decryptMiddleware func(http.Handler) http.Handler
	if privateKey != nil {
		decryptMiddleware = createDecryptMiddleware(privateKey)
	}

	// Создаем роутер с помощью gorilla/mux
	router := mux.NewRouter()

	// Добавляем middleware в правильном порядке:
	// 1. Сначала распаковка запросов (GzipDecompress)
	// 2. Дешифрование (если есть ключ)
	// 3. Затем сжатие ответов (GzipCompress)
	// 4. В конце логирование (WithLogging)
	router.Use(middleware.GzipDecompress)
	if decryptMiddleware != nil {
		router.Use(decryptMiddleware)
	}
	router.Use(middleware.GzipCompress)
	router.Use(WithLogging)

	// Регистрируем маршруты
	// Сначала регистрируем JSON эндпоинты (более специфичные по методу)
	// Используем Path() для точного сопоставления
	router.Path("/update").Handler(updateJSONHandler).Methods("POST")
	router.Path("/update/").Handler(updateJSONHandler).Methods("POST")
	router.Path("/updates").Handler(updatesHandler).Methods("POST")
	router.Path("/updates/").Handler(updatesHandler).Methods("POST")
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
		"POST /updates", "add metrics batch (JSON)",
		"GET /value/{type}/{name}", "get metric value (URL params)",
		"POST /value", "get metric value (JSON)",
		"GET /ping", "database ping",
		"GET /", "metrics dashboard",
	)

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Канал для ошибок сервера
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Ожидаем сигнал остановки
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Ожидаем либо сигнал, либо ошибку сервера
	select {
	case err := <-errCh:
		sugar.Errorw(err.Error(), "event", "server error")
	case <-sigCh:
		sugar.Infow("Shutting down server...")
	}

	// Graceful shutdown с таймаутом 10 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Сначала останавливаем сервер, чтобы не принимать новые запросы
	if err := srv.Shutdown(ctx); err != nil {
		sugar.Errorw("Server forced to shutdown", "error", err)
	}

	// Теперь сохраняем метрики (сервер уже не принимает запросы)
	if storagePath != "" && databaseDSN == "" {
		sugar.Infow("Saving metrics before shutdown...")
		if memStorage, ok := finalStorage.(*storage.MemStorage); ok {
			if err := storage.SaveMetrics(memStorage, storagePath); err != nil {
				sugar.Errorw("Failed to save metrics on shutdown", "error", err)
			} else {
				sugar.Infow("Metrics saved successfully")
			}
		}
	}

	// Закрываем БД, если используется
	if db != nil {
		db.Close()
	}

	sugar.Infow("Server stopped")
}
