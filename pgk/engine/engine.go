package engine 

import "sort"

// IdentifyGhosts now takes metrics from Dashboards, Rules, and Query Logs.
func IdentifyGhosts(allProm []string, grafana map[string][]string, rules map[string]bool, logs map[string]bool) []string {
	// 1. Create a Master Set for O(1) intersection checks
	masterSet := make(map[string]bool)
	for _, m := range allProm {
		masterSet[m] = true
	}

	// 2. The "Verified Used" map (The result of the Intersection)
	verifiedUsed := make(map[string]bool)

	// Intersect Grafana candidates
	for m := range grafana {
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

	// 3. Final Ghost List (Master - VerifiedUsed)
	var ghosts []string
	for _, m := range allProm {
		if !verifiedUsed[m] {
			ghosts = append(ghosts, m)
		}
	}

	sort.Strings(ghosts)
	return ghosts
}