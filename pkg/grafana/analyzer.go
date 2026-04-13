// Copyright 2026 dominikhei
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grafana

import (
	"regexp"
	"sync"
)

// MetricRegex matches Prometheus-style metric names.
var MetricRegex = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_:]+`)

// Analyzer contains the Grafana client from client.go
type Analyzer struct {
	client *Client
}

func NewAnalyzer(client *Client) *Analyzer {
	return &Analyzer{client: client}
}

// DiscoverUsedMetrics crawls all dashboards and returns a unique set of metric names found.
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
