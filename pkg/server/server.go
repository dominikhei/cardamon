package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dominikhei/cardamon/pkg/audit"
	"github.com/dominikhei/cardamon/pkg/droprules"
)

// Server provides an HTTP interface to inspect unused ("ghost") metrics
// and generate drop rules for them.
//
// It serves a simple dashboard UI and JSON APIs for retrieving ghost
// metrics and computing Prometheus drop rules grouped by job.
type Server struct {
	ghosts []audit.MetricReport
	mux    *http.ServeMux
}

// New creates and initializes a Server instance.
//
// It registers HTTP handlers for:
//   - "/"              : serves the dashboard HTML UI
//   - "/api/ghosts"    : returns all ghost metrics as JSON
//   - "/api/droprules" : accepts a list of metric names and returns
//     generated drop rules grouped by job
//
// The provided ghosts slice is used as the data source for all endpoints.
func New(ghosts []audit.MetricReport) *Server {
	mux := http.NewServeMux()
	s := &Server{
		ghosts: ghosts,
		mux:    mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(dashboardHTML)
	})

	mux.HandleFunc("/api/ghosts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(s.ghosts); err != nil {
			http.Error(w, "failed to encode json", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/droprules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var names []string
		if err := json.NewDecoder(r.Body).Decode(&names); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		jobByName := make(map[string]string)
		for _, g := range s.ghosts {
			jobByName[g.Name] = g.Job
		}

		jobMap := make(map[string][]string)
		for _, name := range names {
			job := jobByName[name]
			jobMap[job] = append(jobMap[job], name)
		}

		type JobRules struct {
			Job   string           `json:"job"`
			Rules []droprules.Rule `json:"rules"`
		}

		var result []JobRules
		for job, metrics := range jobMap {
			result = append(result, JobRules{
				Job:   job,
				Rules: droprules.Generate(metrics),
			})
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, "failed to encode json", http.StatusInternalServerError)
		}
	})

	return s
}

func (s *Server) ListenAndServe(addr string) error {
	log.Printf("dashboard: http://localhost%v/\n", addr)
	return http.ListenAndServe(addr, s.mux)
}
