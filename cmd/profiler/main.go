package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	serverAddr      = "localhost:8080"
	profileURL      = "http://" + serverAddr + "/debug/pprof/heap"
	updatesURL      = "http://" + serverAddr + "/updates"
	valueURL        = "http://" + serverAddr + "/value/gauge/metric_0"
	indexURL        = "http://" + serverAddr + "/"
	profileDuration = 30 * time.Second
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: profiler <output_file>")
		os.Exit(1)
	}
	outputFile := os.Args[1]

	// Запускаем сервер в фоне
	serverCmd := exec.Command("go", "run", "cmd/server/main.go", "-a", serverAddr)
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	if err := serverCmd.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer serverCmd.Process.Kill()

	// Ждём, пока сервер запустится
	time.Sleep(2 * time.Second)

	// Генерируем нагрузку
	fmt.Println("Generating load...")
	generateLoad()

	// Ждём, пока память заполнится
	time.Sleep(2 * time.Second)

	// Снимаем профиль
	fmt.Println("Collecting profile...")
	if err := collectProfile(outputFile); err != nil {
		log.Fatalf("Failed to collect profile: %v", err)
	}

	fmt.Printf("Profile saved to %s\n", outputFile)
}

func generateLoad() {
	// Создаём метрики для отправки - 100 метрик в батче
	var b strings.Builder
	b.WriteString(`[`)
	for i := 0; i < 100; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(fmt.Sprintf(`{"id":"metric_%d","type":"gauge","value":%d.0}`, i, i))
	}
	b.WriteString(`]`)
	metricsJSON := b.String()

	client := &http.Client{}

	// Отправляем много запросов с батчами
	for i := 0; i < 500; i++ {
		resp, err := client.Post(updatesURL, "application/json", bytes.NewBufferString(metricsJSON))
		if err != nil {
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		// Добавляем GET запросы для чтения
		if i%5 == 0 {
			resp, err := client.Get(indexURL)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}

		// Отправляем одиночные метрики
		if i%3 == 0 {
			resp, err := http.Post(serverAddr+"/update/gauge/metric_"+fmt.Sprintf("%d", i%100)+"/1.5", "text/plain", nil)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}

		// Читаем значения
		if i%10 == 0 {
			resp, err := client.Get(valueURL)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}
	}
}

func collectProfile(outputFile string) error {
	// Создаём файл для профиля
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Запрашиваем профиль памяти
	client := &http.Client{Timeout: profileDuration}
	resp, err := client.Get(profileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Копируем данные в файл
	_, err = io.Copy(file, resp.Body)
	return err
}
