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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dominikhei/cardamon/pkg/config"
	"github.com/dominikhei/cardamon/pkg/engine"
	"github.com/dominikhei/cardamon/pkg/grafana"
	"github.com/dominikhei/cardamon/pkg/prom"
	"github.com/dominikhei/cardamon/pkg/server"
)

func main() {

	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	promClient, err := prom.NewClient(cfg.Prometheus.Address+cfg.Prometheus.PathPrefix, cfg.Prometheus.Token, cfg.Prometheus.Username, cfg.Prometheus.Password)
	if err != nil {
		log.Fatalf("Failed to initialize Prometheus client: %v", err)
	}
	promAnalyzer := prom.NewAnalyzer(promClient)
	grafanaClient := grafana.NewClient(cfg.Grafana.Address, cfg.Grafana.PathPrefix, cfg.Grafana.Token, cfg.Grafana.Username, cfg.Grafana.Password)
	grafanaAnalyzer := grafana.NewAnalyzer(grafanaClient)

	ctx := context.Background()

	allMetrics, err := promAnalyzer.GetAllMetricNames(ctx)
	if err != nil {
		log.Fatalf("Error fetching all metric names from Prometheus: %v", err)
	}

	grafanaUsed, err := grafanaAnalyzer.DiscoverUsedMetrics(ctx)
	if err != nil {
		fmt.Printf("Warning: Creawling Grafana failed: %v\n", err)
	}

	rulesUsed, err := promAnalyzer.GetMetricsInRules(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch Prometheus rules: %v\n", err)
	}

	var logsUsed map[string]bool
	if cfg.Storage.QueryLogDir != "" {

		logsUsed, err = promAnalyzer.DiscoverUsedMetricsFromLogs(
			cfg.Storage.QueryLogDir,
			cfg.Storage.LookbackDays,
		)
		if err != nil {
			fmt.Printf("Warning: Log scan failed: %v\n", err)
		}
	}
	allMetrics = promAnalyzer.FilterMetrics(allMetrics, cfg.Audit.ExcludePrefixes)
	ghosts := engine.IdentifyGhosts(allMetrics, grafanaUsed, rulesUsed, logsUsed)
	ghostReports, err := promAnalyzer.GetGhostStats(ctx, ghosts)
	if err != nil {
		log.Fatalf("Failed to fetch ghost stats: %v", err)
	}

	apiGhosts := make([]prom.MetricReport, 0, len(ghostReports))
	for _, g := range ghostReports {
		apiGhosts = append(apiGhosts, prom.MetricReport{
			Name:             g.Name,
			Job:              g.Job,
			SeriesCount:      g.SeriesCount,
			LabelCount:       g.LabelCount,
			InactiveDuration: g.InactiveDuration,
		})
	}
	addr := fmt.Sprintf(":%d", cfg.Dashboard.Port)
	srv := server.New(apiGhosts)
	log.Fatal(srv.ListenAndServe(addr))

}
