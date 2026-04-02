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
)

func main() {

	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	promClient, err := prom.NewClient(cfg.Prometheus.Address)
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

	ghosts := engine.IdentifyGhosts(allMetrics, grafanaUsed, rulesUsed, logsUsed)

	fmt.Println("\n--- AUDIT SUMMARY ---")
	fmt.Printf("✅ Total Active Metrics:    %d\n", len(allMetrics))
	fmt.Printf("👻 Total Ghost Metrics:     %d\n", len(ghosts))

	if len(allMetrics) > 0 {
		efficiency := float64(len(allMetrics)-len(ghosts)) / float64(len(allMetrics)) * 100
		fmt.Printf("📈 Utilization Score:      %.2f%%\n", efficiency)
	}
	
}