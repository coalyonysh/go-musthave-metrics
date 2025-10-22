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

	// Получаем все метрики
	gauges := h.storage.GetAllGauges()
	counters := h.storage.GetAllCounters()

	// Создаем HTML страницу
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Metrics Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        h2 { color: #666; margin-top: 30px; }
        table { border-collapse: collapse; width: 100%; margin-top: 10px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .metric-name { font-family: monospace; }
        .metric-value { font-family: monospace; }
        .no-metrics { color: #999; font-style: italic; }
    </style>
</head>
<body>
    <h1>Metrics Server</h1>
    <p>Сервер сбора метрик работает!</p>
    
    <h2>Gauge метрики</h2>`

	if len(gauges) == 0 {
		html += `<p class="no-metrics">Нет gauge метрик</p>`
	} else {
		html += `<table>
        <tr>
            <th>Имя метрики</th>
            <th>Значение</th>
        </tr>`

		// Сортируем ключи для консистентного отображения
		var gaugeNames []string
		for name := range gauges {
			gaugeNames = append(gaugeNames, name)
		}
		sort.Strings(gaugeNames)

		for _, name := range gaugeNames {
			value := gauges[name]
			html += fmt.Sprintf(`
        <tr>
            <td class="metric-name">%s</td>
            <td class="metric-value">%g</td>
        </tr>`, name, value)
		}
		html += `</table>`
	}

	html += `
    <h2>Counter метрики</h2>`

	if len(counters) == 0 {
		html += `<p class="no-metrics">Нет counter метрик</p>`
	} else {
		html += `<table>
        <tr>
            <th>Имя метрики</th>
            <th>Значение</th>
        </tr>`

		// Сортируем ключи для консистентного отображения
		var counterNames []string
		for name := range counters {
			counterNames = append(counterNames, name)
		}
		sort.Strings(counterNames)

		for _, name := range counterNames {
			value := counters[name]
			html += fmt.Sprintf(`
        <tr>
            <td class="metric-name">%s</td>
            <td class="metric-value">%d</td>
        </tr>`, name, value)
		}
		html += `</table>`
	}

	html += `
    <h2>API</h2>
    <p>Используйте следующие эндпоинты:</p>
    <ul>
        <li><strong>POST</strong> /update/&lt;type&gt;/&lt;name&gt;/&lt;value&gt; - добавить метрику</li>
        <li><strong>GET</strong> /value/&lt;type&gt;/&lt;name&gt; - получить значение метрики</li>
        <li><strong>GET</strong> / - эта страница</li>
    </ul>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
