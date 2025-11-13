package handlers

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/coalyonysh/go-musthave-metrics/internal/storage"
)

type IndexHandler struct {
	storage storage.Storage
}

func NewIndexHandler(storage storage.Storage) *IndexHandler {
	return &IndexHandler{
		storage: storage,
	}
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Устанавливаем Content-Type для правильной работы gzip middleware
	w.Header().Set("Content-Type", "text/html")

	gauges := h.storage.GetAllGauges()
	counters := h.storage.GetAllCounters()

	// Простая HTML-страница со списком метрик
	fmt.Fprintf(w, "<html><body>")
	fmt.Fprintf(w, "<h1>Metrics</h1>")

	fmt.Fprintf(w, "<h2>Gauge</h2>")
	if len(gauges) == 0 {
		fmt.Fprintf(w, "<p>нет gauge</p>")
	} else {
		var names []string
		for n := range gauges {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			fmt.Fprintf(w, "%s: %g<br>", n, gauges[n])
		}
	}

	fmt.Fprintf(w, "<h2>Counter</h2>")
	if len(counters) == 0 {
		fmt.Fprintf(w, "<p>нет counter</p>")
	} else {
		var names []string
		for n := range counters {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			fmt.Fprintf(w, "%s: %d<br>", n, counters[n])
		}
	}

	fmt.Fprintf(w, "</body></html>")
}
