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
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"
)

// seenRequest is what the fake Prometheus API received.
type seenRequest struct {
	path   string
	tenant []string
	auth   string
}

func TestNewClient_TenantHeader(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		tenantID   string
		wantTenant []string
		wantAuth   string
	}{
		{name: "no tenant sends no header"},
		{name: "tenant is sent", tenantID: "team-a", wantTenant: []string{"team-a"}},
		{name: "value is sent unchanged", tenantID: "team-a|team-b", wantTenant: []string{"team-a|team-b"}},
		{name: "tenant is sent alongside auth", token: "secret", tenantID: "team-a", wantTenant: []string{"team-a"}, wantAuth: "Bearer secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mu   sync.Mutex
				seen []seenRequest
			)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				seen = append(seen, seenRequest{
					path:   r.URL.Path,
					tenant: r.Header.Values("X-Scope-OrgID"),
					auth:   r.Header.Get("Authorization"),
				})
				mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"status":"success","data":["up","http_requests_total"]}`))
			}))
			defer srv.Close()

			// Mimir serves the Prometheus API under /prometheus, which cardamon gets via path_prefix.
			c, err := NewClient(srv.URL+"/prometheus", tt.token, "", "", tt.tenantID)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			names, err := NewAnalyzer(c).GetAllMetricNames(ctx)
			if err != nil {
				t.Fatalf("GetAllMetricNames: %v", err)
			}
			if want := []string{"up", "http_requests_total"}; !slices.Equal(names, want) {
				t.Errorf("metric names = %v, want %v", names, want)
			}

			mu.Lock()
			defer mu.Unlock()
			if len(seen) != 1 {
				t.Fatalf("got %d requests, want 1", len(seen))
			}
			got := seen[0]
			if want := "/prometheus/api/v1/label/__name__/values"; got.path != want {
				t.Errorf("path = %q, want %q", got.path, want)
			}
			if !slices.Equal(got.tenant, tt.wantTenant) {
				t.Errorf("X-Scope-OrgID = %q, want %q", got.tenant, tt.wantTenant)
			}
			if got.auth != tt.wantAuth {
				t.Errorf("Authorization = %q, want %q", got.auth, tt.wantAuth)
			}
		})
	}
}
