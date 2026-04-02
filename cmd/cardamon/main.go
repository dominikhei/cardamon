package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dominikhei/cardamon/pgk/config"
	"github.com/dominikhei/cardamon/pgk/engine"
	"github.com/dominikhei/cardamon/pgk/grafana"
	"github.com/dominikhei/cardamon/pgk/prom"
	"github.com/dominikhei/cardamon/pgk/server"
)

func main() {

	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	promClient, err := prom.NewClient(cfg.Prometheus.Address + cfg.Prometheus.PathPrefix)
	if err != nil {
		log.Fatalf("Failed to initialize Prometheus client: %v", err)
	}
	promAnalyzer := prom.NewAnalyzer(promClient)
	if err != nil {
		log.Fatalf("Failed to initialize analyzer: %v", err)
	}
	grafanaClient := grafana.NewClient(cfg.Grafana.Address, cfg.Grafana.PathPrefix, cfg.Grafana.ApiKey)
	grafanaAnalyzer := grafana.NewAnalyzer(grafanaClient)

	ctx := context.Background()

	allMetrics, err := promAnalyzer.GetAllMetricNames(ctx)
	if err != nil {
		log.Fatalf("Error fetching master list: %v", err)
	}

	grafanaUsed, err := grafanaAnalyzer.DiscoverUsedMetrics()
	if err != nil {
		fmt.Printf("Warning: Grafana crawl failed: %v\n", err)
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
	ghostReports, err :=  promAnalyzer.GetGhostStats(ctx, ghosts)
	if err != nil {
		log.Fatalf("Failed to fetch ghost stats: %v", err)
	}



	apiGhosts := make([]server.GhostMetric, 0, len(ghostReports))
	for _, g := range ghostReports {
		apiGhosts = append(apiGhosts, server.GhostMetric{
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