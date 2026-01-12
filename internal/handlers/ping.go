package handlers

import (
	"database/sql"
	"net/http"
)

type PingHandler struct {
	db *sql.DB
}

func NewPingHandler(db *sql.DB) *PingHandler {
	return &PingHandler{
		db: db,
	}
}

func (h *PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	if err := h.db.Ping(); err != nil {
		http.Error(w, "Database ping failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
