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
func (a *Analyzer) DiscoverUsedMetrics() ([]string, error) {
    dashboards, err := a.client.SearchDashboards()
    if err != nil {
        return nil, err
    }

    seen := make(map[string]bool)
    var mu sync.Mutex
    var wg sync.WaitGroup
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

            matches := MetricRegex.FindAllString(string(rawJSON), -1)
            mu.Lock()
            for _, m := range matches {
                seen[m] = true
            }
            mu.Unlock()
        }(dash)
    }

    wg.Wait()

    metrics := make([]string, 0, len(seen))
    for k := range seen {
        metrics = append(metrics, k)
    }
    return metrics, nil
}