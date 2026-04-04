// Copyright 2026 Google LLC
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

package audit

// MetricReport contains all relevant data for a metric that is required in the dashboard
type MetricReport struct {
	Name             string `json:"name"`
	Job              string `json:"job"`
	SeriesCount      int    `json:"series_count"`
	LabelCount       int    `json:"label_count"`
	InactiveDuration string `json:"inactive_duration"`
}
