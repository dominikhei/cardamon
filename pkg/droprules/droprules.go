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

package droprules

import (
	"sort"
	"strings"
)

// Rule contains a droprule regex and the affected metrics.
type Rule struct {
	Regex   string   `json:"regex"`
	Metrics []string `json:"metrics"`
}

// Generate generates a drop rule for a set of metrics.
func Generate(metrics []string) []Rule {
	sort.Strings(metrics)
	groups := make(map[string][]string)

	for _, m := range metrics {
		prefix := findBestPrefix(m, metrics)
		groups[prefix] = append(groups[prefix], m)
	}

	var rules []Rule
	for prefix, group := range groups {
		if len(group) == 1 {
			rules = append(rules, Rule{
				Regex:   group[0],
				Metrics: group,
			})
			continue
		}

		suffixes := make([]string, len(group))
		for i, m := range group {
			suffixes[i] = strings.TrimPrefix(m, prefix+"_")
		}
		sort.Strings(suffixes)

		rules = append(rules, Rule{
			Regex:   prefix + "_(" + strings.Join(suffixes, "|") + ")",
			Metrics: group,
		})
	}

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Regex < rules[j].Regex
	})
	return rules
}

// findBestPrefix checks the longest matching prefix for a set of metrics to generate a drop rule for multiple metrics.
func findBestPrefix(metric string, all []string) string {
	parts := strings.Split(metric, "_")
	best := parts[0]
	for i := 2; i < len(parts); i++ {
		candidate := strings.Join(parts[:i], "_")
		for _, m := range all {
			if m != metric && strings.HasPrefix(m, candidate+"_") {
				best = candidate
				break
			}
		}
	}
	return best
}
