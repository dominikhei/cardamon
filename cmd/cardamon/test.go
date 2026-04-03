package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// MetricRegex matches strings that look like Prometheus metrics
// Starts with letter/underscore, followed by alphanumeric/colon/underscore
var MetricRegex = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_:]+`)

// Common PromQL keywords to ignore
var skipList = map[string]bool{
	"sum": true, "avg": true, "min": true, "max": true, "count": true,
	"rate": true, "irate": true, "delta": true, "increase": true,
	"histogram_quantile": true, "by": true, "without": true, "on": true,
	"ignoring": true, "group_left": true, "group_right": true, "bool": true,
}

func main() {
	// 1. SET YOUR LOCAL FILE PATH HERE
	filePath := "dashboard.json"

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ Error reading file: %v\n", err)
		return
	}

	// 2. Run the Regex across the entire raw JSON blob
	matches := MetricRegex.FindAllString(string(content), -1)

	// 3. Deduplicate and Filter
	uniqueMetrics := make(map[string]bool)
	for _, m := range matches {
		// Clean up: lowercase check for keywords
		lowered := strings.ToLower(m)
		
		// Rules for a "Good" candidate:
		// - Not in the skip list
		// - Longer than 3 chars (prevents catching 'id', 'x', etc.)
		// - Doesn't start with a number (handled by regex, but good to be safe)
		if !skipList[lowered] && len(m) > 3 {
			uniqueMetrics[m] = true
		}
	}

	// 4. Sort for pretty output
	var result []string
	for m := range uniqueMetrics {
		result = append(result, m)
	}
	sort.Strings(result)

	// 5. Print Results
	fmt.Printf("📊 Found %d unique metric candidates in %s:\n", len(result), filePath)
	fmt.Println(strings.Repeat("-", 40))
	for _, m := range result {
		fmt.Printf("  • %s\n", m)
	}
}