package engine

import (
	"sort"
)

// IdentifyGhosts takes metrics from Dashboards, Rules, and Query Logs and uses set operation to determine unused metrics.
// The formula is: Ghost Metrics = All Metrics - (Grafana Dashboard Metrics + Prometheus Rules + Metrics queried in the last n days)
func IdentifyGhosts(allProm []string, grafana []string, rules map[string]bool, logs map[string]bool) []string {
	masterSet := make(map[string]bool)
	for _, m := range allProm {
		masterSet[m] = true
	}

	verifiedUsed := make(map[string]bool)

	// Intersect Grafana candidates
	for _, m := range grafana {
		if masterSet[m] {
			verifiedUsed[m] = true
		}
	}

	// Intersect Rule candidates
	for m := range rules {
		if masterSet[m] {
			verifiedUsed[m] = true
		}
	}

	// Intersect Query Log candidates
	for m := range logs {
		if masterSet[m] {
			verifiedUsed[m] = true
		}
	}

	var ghosts []string
	for _, m := range allProm {
		if !verifiedUsed[m] {
			ghosts = append(ghosts, m)
		}
	}
	sort.Strings(ghosts)
	return ghosts
}
