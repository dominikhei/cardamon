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
	"context"
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
func (a *Analyzer) DiscoverUsedMetrics(ctx context.Context) ([]string, error) {
	dashboards, err := a.client.SearchDashboards(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for _, dash := range dashboards {
		// The Semaphore limits the amount of goroutines that can be created and thus also the API calls
		sem <- struct{}{}
		wg.Add(1)
		go func(d DashboardMetadata) {
			defer wg.Done()
			defer func() { <-sem }()

			rawJSON, err := a.client.GetDashboardModel(ctx, d.UID)
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
