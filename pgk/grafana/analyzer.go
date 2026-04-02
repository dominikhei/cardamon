package grafana

import (
	"regexp"
	"sync"
)

// MetricRegex matches Prometheus-style metric names. 
// It looks for sequences of characters that start with a letter/underscore 
// and contain alphanumeric/colons/underscores.
var MetricRegex = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_:]+`)

type Analyzer struct {
	client *Client
}

func NewAnalyzer(client *Client) *Analyzer {
	return &Analyzer{client: client}
}

// DiscoverUsedMetrics crawls all dashboards and returns a unique set of metric names found
func (a *Analyzer) DiscoverUsedMetrics() (map[string][]string, error) {
	dashboards, err := a.client.SearchDashboards()
	if err != nil {
		return nil, err
	}

	usedMetrics := make(map[string][]string) // metric -> []dashboardTitles
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore to limit concurrency (don't DDoS your own Grafana)
	sem := make(chan struct{}, 10)

	for _, dash := range dashboards {
		wg.Add(1)
		go func(d DashboardMetadata) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			rawJSON, err := a.client.GetDashboardModel(d.UID)
			if err != nil {
				return
			}

			// Find all potential metric strings in the JSON
			matches := MetricRegex.FindAllString(string(rawJSON), -1)

			mu.Lock()
			for _, m := range matches {
				// Avoid duplicates within the same dashboard list
				if !contains(usedMetrics[m], d.Title) {
					usedMetrics[m] = append(usedMetrics[m], d.Title)
				}
			}
			mu.Unlock()
		}(dash)
	}

	wg.Wait()
	return usedMetrics, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}