package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dominikhei/cardamon/pgk/engine"
	"github.com/dominikhei/cardamon/pgk/config"
	"github.com/dominikhei/cardamon/pgk/prom"
)

func main() {
	// 1. Define the CLI flag for the configuration file
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// 2. Load and validate the YAML configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("❌ Configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🚀 Cardi: The Prometheus Ghost Finder")
	fmt.Println("------------------------------------")

	// 3. Initialize the Analyzer (The core engine for API communication)
	// Passing Prometheus Addr, Grafana Addr, and Grafana Key from config
	analyzer := prom.NewAnalyzer(
		cfg.Prometheus.Address,
	)
	if err != nil {
		log.Fatalf("❌ Failed to initialize analyzer: %v", err)
	}

	ctx := context.Background()

	// 4. Step 1: Get the Master List (All metrics currently in Prometheus)
	fmt.Printf("📥 Fetching master metric list from %s... ", cfg.Prometheus.Address)
	allMetrics, err := analyzer.GetAllMetricNames(ctx)
	if err != nil {
		log.Fatalf("\n❌ Error fetching master list: %v", err)
	}
	fmt.Printf("Done (%d metrics found).\n", len(allMetrics))

	// 5. Step 2: Crawl Grafana Dashboards
	fmt.Printf("📊 Crawling Grafana dashboards at %s... ", cfg.Grafana.Address)
	grafanaUsed, err := analyzer.DiscoverUsedMetrics()
	if err != nil {
		fmt.Printf("\n⚠️  Warning: Grafana crawl partially failed: %v\n", err)
	} else {
		fmt.Printf("Done (%d unique metrics referenced).\n", len(grafanaUsed))
	}

	// 6. Step 3: Fetch Prometheus Alerting/Recording Rules
	fmt.Print("🚨 Fetching Prometheus alerting and recording rules... ")
	rulesUsed, err := analyzer.GetMetricsInRules(ctx)
	if err != nil {
		fmt.Printf("\n⚠️  Warning: Failed to fetch rules: %v\n", err)
	} else {
		fmt.Printf("Done (%d metrics used in rules).\n", len(rulesUsed))
	}

	// 7. Step 4: Scan Local Query Logs (If directory is provided)
	var logsUsed map[string]bool
	if cfg.Storage.QueryLogDir != "" {
		fmt.Printf("📂 Scanning local query logs in %s (Lookback: %d days)... ", 
			cfg.Storage.QueryLogDir, cfg.Storage.LookbackDays)
		
		logsUsed, err = analyzer.DiscoverUsedMetricsFromLogs(
			cfg.Storage.QueryLogDir, 
			cfg.Storage.LookbackDays,
		)
		if err != nil {
			fmt.Printf("\n⚠️  Warning: Log scan failed: %v\n", err)
		} else {
			fmt.Printf("Done (%d metrics found in active use).\n", len(logsUsed))
		}
	} else {
		fmt.Println("⏭️  Skipping Query Logs (no local path provided in config).")
	}

	// 8. Step 5: Execute the Audit Engine (The Math)
	// This identifies metrics that exist in Prometheus but are NOT in Dashboards, Rules, or Logs.
	fmt.Println("🔍 Analyzing intersections to identify Ghosts...")
	ghosts := audit.IdentifyGhosts(allMetrics, grafanaUsed, rulesUsed, logsUsed)

	// 9. Final Results & Output
	fmt.Println("\n--- AUDIT SUMMARY ---")
	fmt.Printf("✅ Total Active Metrics:    %d\n", len(allMetrics))
	fmt.Printf("👻 Total Ghost Metrics:     %d\n", len(ghosts))

	if len(allMetrics) > 0 {
		efficiency := float64(len(allMetrics)-len(ghosts)) / float64(len(allMetrics)) * 100
		fmt.Printf("📈 Utilization Score:      %.2f%%\n", efficiency)
	}

	// 10. Save Results
	if len(ghosts) > 0 {
		err := audit.SaveReport(cfg.Output.File, ghosts)
		if err != nil {
			log.Fatalf("❌ Failed to save report: %v", err)
		}
		fmt.Printf("\n💾 Ghost list saved to: %s\n", cfg.Output.File)
		fmt.Println("💡 Tip: Review these metrics. If they aren't needed, use 'metric_relabel_configs' to drop them.")
	} else {
		fmt.Println("\n✨ No ghosts found! Your Prometheus instance is perfectly utilized.")
	}
}