package prom

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	v1 "github.com/prometheus/prometheus/web/api/v1"
)

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

// Reuse the same regex logic from Grafana to find metrics inside Alert expressions
var MetricRegex = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_:]+`)

// GetMetricsInRules fetches all metrics currently used in Alerts and Recording Rules
func (a *Analyzer) GetMetricsInRules(ctx context.Context) (map[string]bool, error) {
	ruleGroups, err := a.client.api.Rules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prometheus rules: %w", err)
	}

	usedInRules := make(map[string]bool)

	for _, group := range ruleGroups.Groups {
		for _, rule := range group.Rules {
			var query string

			// Handle both Alerting and Recording rules
			switch r := rule.(type) {
			case v1.AlertingRule:
				query = r.Query
			case v1.RecordingRule:
				query = r.Query
			}

			// Extract metric names from the PromQL expression
			matches := MetricRegex.FindAllString(query, -1)
			for _, m := range matches {
				// We'll filter these against the Master List later in the Audit stage
				usedInRules[m] = true
			}
		}
	}

	return usedInRules, nil
}

type QueryLogEntry struct {
	Params struct {
		Query string `json:"query"`
	} `json:"params"`
}

// DiscoverUsedMetricsFromLogs scans a directory for Prometheus query logs.
// It handles both active .log files and rotated .gz files.
func (a *Analyzer) DiscoverUsedMetricsFromLogs(logDir string, days int) (map[string]bool, error) {
	usedInLogs := make(map[string]bool)

	// Calculate the cutoff time
	cutoff := time.Now().AddDate(0, 0, -days)

	files, err := os.ReadDir(logDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !strings.Contains(file.Name(), "query.log") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Skip files older than our flag limit
		if info.ModTime().Before(cutoff) {
			continue
		}

		path := filepath.Join(logDir, file.Name())
		err = a.parseLogFile(path, usedInLogs)
		if err != nil {
			continue
		}
	}

	return usedInLogs, nil
}

func (a *Analyzer) parseLogFile(path string, found map[string]bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer gz.Close()
		reader = gz
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var entry QueryLogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			// Reuse your existing MetricRegex
			matches := MetricRegex.FindAllString(entry.Params.Query, -1)
			for _, m := range matches {
				found[m] = true
			}
		}
	}
	return scanner.Err()
}
