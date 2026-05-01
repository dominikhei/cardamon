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

package prom

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/common/model"
)

// Analyzer contains the Prometheus client from client.go.
type Analyzer struct {
	client *Client
}

func NewAnalyzer(client *Client) *Analyzer {
	return &Analyzer{client: client}
}

// GetAllMetricNames returns every metric name currently in the Prometheus index.
// We look back 24h to ensure we don't miss infrequent batch jobs.
func (a *Analyzer) GetAllMetricNames(ctx context.Context) ([]string, error) {
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	// Fetch all values for the label "__name__"
	labelValues, _, err := a.client.api.LabelValues(ctx, "__name__", []string{}, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("prometheus label_values failed: %w", err)
	}

	metrics := make([]string, len(labelValues))
	for i, v := range labelValues {
		metrics[i] = string(v)
	}
	return metrics, nil
}

// Regex to find metrics inside Alert expressions.
var MetricRegex = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_:]+`)

// GetMetricsInRules fetches all metrics currently used in Alerts and Recording Rules.
func (a *Analyzer) GetMetricsInRules(ctx context.Context) (map[string]bool, error) {
	ruleGroups, err := a.client.api.Rules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prometheus rules: %w", err)
	}

	usedInRules := make(map[string]bool)

	for _, group := range ruleGroups.Groups {
		for _, rule := range group.Rules {
			raw := fmt.Sprintf("%+v", rule)
			matches := MetricRegex.FindAllString(raw, -1)
			for _, m := range matches {
				usedInRules[m] = true
			}
		}
	}

	return usedInRules, nil
}

// FilterMetrics filters metrics based on an exclusion list.
func (a *Analyzer) FilterMetrics(metrics []string, excludePrefixes []string) []string {
	var filtered []string
	for _, m := range metrics {
		excluded := false
		for _, prefix := range excludePrefixes {
			if strings.HasPrefix(m, strings.TrimSpace(prefix)) {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// MetricReport contains all relevant data for a metric that is required in the dashboard
type MetricReport struct {
	Name             string `json:"name"`
	Job              string `json:"job"`
	SeriesCount      int    `json:"series_count"`
	LabelCount       int    `json:"label_count"`
	InactiveDuration string `json:"inactive_duration"`
}

// GetGhostStats calculates general statistics, like the series count, label count and when it was last scraped.
func (a *Analyzer) GetGhostStats(ctx context.Context, ghosts []string) ([]MetricReport, error) {
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	reports := make([]MetricReport, len(ghosts))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for i, name := range ghosts {
		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			report := MetricReport{Name: name}

			series, _, err := a.client.api.Series(ctx, []string{name}, startTime, endTime)
			if err == nil {
				report.SeriesCount = len(series)

				labelSet := make(map[string]bool)
				for _, s := range series {
					for k := range s {
						labelSet[string(k)] = true
					}
				}
				report.LabelCount = len(labelSet)

				if len(series) > 0 {
					report.Job = string(series[0]["job"])
				}
			}

			result, _, err := a.client.api.Query(ctx, fmt.Sprintf("timestamp(%s)", name), time.Now())
			if err == nil {
				if vec, ok := result.(model.Vector); ok && len(vec) > 0 {
					lastReceived := time.Unix(int64(vec[0].Value), 0)
					report.InactiveDuration = formatDuration(time.Since(lastReceived))
				}
			}

			reports[i] = report
		}(i, name)
	}

	wg.Wait()
	return reports, nil
}

// Helper function to format the inactive duration for the UI.
func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
