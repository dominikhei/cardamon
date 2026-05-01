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
	"bufio"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// QueryLogEntry contains a query from the query log.
type QueryLogEntry struct {
	Time   time.Time `json:"time"`
	Params struct {
		Query string `json:"query"`
	} `json:"params"`
}

func (a *Analyzer) DiscoverUsedMetricsFromLogs(logDir string, days int) (map[string]bool, error) {
	usedInLogs := make(map[string]bool)
	cutoff := time.Now().AddDate(0, 0, -days)

	files, err := os.ReadDir(logDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		name := file.Name()
		if !strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".log.gz") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Skip rotated files that are entirely too old
		if info.ModTime().Before(cutoff) {
			continue
		}

		path := filepath.Join(logDir, name)
		if err = a.parseLogFile(path, usedInLogs, cutoff); err != nil {
			continue
		}
	}

	return usedInLogs, nil
}

// parseLogFile scans a logFile for all expressions that are potentially metrics.
// It breaks early once it hits entries older than the cutoff.
func (a *Analyzer) parseLogFile(path string, found map[string]bool, cutoff time.Time) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	var reader io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer gz.Close() //nolint:errcheck
		reader = gz
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var entry QueryLogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		if entry.Time.Before(cutoff) {
			break
		}

		matches := MetricRegex.FindAllString(entry.Params.Query, -1)
		for _, m := range matches {
			found[m] = true
		}
	}

	return scanner.Err()
}
