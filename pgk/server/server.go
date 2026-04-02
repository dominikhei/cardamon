package server

import (
	"encoding/json"
	"log"
	"net/http"
)

type GhostMetric struct {
	Name             string `json:"name"`
	Job              string `json:"job,omitempty"`
	SeriesCount      int    `json:"series_count"`
	LabelCount       int    `json:"label_count"`
	InactiveDuration string `json:"inactive_duration"`
}

type Server struct {
	ghosts []GhostMetric
	mux    *http.ServeMux
}

func New(ghosts []GhostMetric) *Server {
	mux := http.NewServeMux()

	s := &Server{
		ghosts: ghosts,
		mux:    mux,
	}

	// Dashboard
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// optional: only allow "/" (avoid surprising 404 vs index)
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(dashboardHTML)
	})

	// API
	mux.HandleFunc("/api/ghosts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(s.ghosts); err != nil {
			http.Error(w, "failed to encode json", http.StatusInternalServerError)
			return
		}
	})

	return s
}

func (s *Server) ListenAndServe(addr string) error {
	log.Printf("dashboard: http://localhost%v/\n", addr)
	return http.ListenAndServe(addr, s.mux)
}