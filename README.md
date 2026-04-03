# Cardamon
**Cardamon** is a metric auditor for Prometheus. It identifies metrics that exist in your TSDB but are never actually queried by dashboards, alerting rules, recording rules, or any other consumer. You can then generate Prometheus drop rules to remove them and reduce storage need.

## Demo
![](./media/cardamon_demo.gif)

## What is an Unused Metric?

An unused metric is a metric that Prometheus scrapes and stores, but that nothing ever reads. It occupies memory, disk, and ingestion budget without providing value. At scale, such metrics can account for a significant portion of your total series count.

Cardamon cross-references every metric in your TSDB against:

- **Prometheus query logs** Every PromQL expression evaluated over a configurable lookback window
- **Prometheus alerting & recording rules** All rule expressions in your Prometheus
- **Grafana dashboards** Every metric across all dashboards in your Grafana instance

Any metric not found in any of those sources is flagged as a ghost.

---

## How It Works

1. Cardamon fetches all metric names from Prometheus using the label values API.
2. It collects the set of metrics that are actively used from three sources: the Prometheus query log, the rules API, and Grafana dashboard JSON.
3. Based on the Overlap with (1) Cardamon identifies actual metrics from (2).
4. The difference between (1) and (3) is your ghost set.
5. For each ghost, Cardamon fetches series count, label cardinality, job, and last-seen timestamp from Prometheus.
6. Results are served via a local web UI where you can explore, filter, and export Prometheus `drop` relabeling rules.

---

## Configuration

Cardamon is configured via a YAML file. Pass the path with `--config`.

```yaml
prometheus:
  address: "http://localhost:9090"
  path_prefix: ""                        # Set if Prometheus is served under a subpath, e.g. /prometheus
                                         # If no auth is configured leave this blank
  token: ""                              # Token based auth, takes precedence over password auth if both are set
  username: ""                           # Basic auth / LDAP fallback
  password: ""

grafana:
  url: "http://localhost:3000"
  path_prefix: ""                        # Set if Grafana is served under a subpath, e.g. /grafana
                                         # If no auth is configured leave this blank
  token: ""                              # Token based auth, takes precedence over password auth if both are set
  username: ""                           # Basic auth / LDAP fallback
  password: ""

storage:                                 # Optional
  query_log_dir: "/var/log/prometheus"   # Directory containing Prometheus query log files
  lookback_days: 7                       # Only consider log entries from the last N days

audit:
  exclude_prefixes:                      # Metric name prefixes to ignore entirely
    - "go_"
    - "process_"
    - "scrape_"

server:
  port: 8080                             # On which port to serve the dashboard
```
**The following keys are required:**
- grafana: address
- prometheus: address
- server: port

### Enabling the Prometheus query log

Cardamon scans the Prometheus query log to discover which metrics are actually being queried.
You have to copy those files to your instance or start cardamon inside the Prometheus server to use them, as cardamon can not query them on another host.
They [need to be enabled](https://prometheus.io/docs/guides/query-log/), e.g via your values.yaml.

> **Note:** Cardamon handles both active `.log` files and rotated `.log.gz` files. It reads log entries in order and stops early once it reaches entries older than `lookback_days`, so scanning is efficient even on large logs.

---

## Usage

```bash
cardamon --config config.yaml
```

Once started, open your browser at the port you configured.

---

## The UI

The UI gives you a full picture of your unused metrics and lets you act on them.

### Metric table

Each row represents one unused metric with the following columns:

| Column | Description |
|---|---|
| **Metric** | The metric name |
| **Job** | The scrape job the metric originates from |
| **Series** | Number of active time series for this metric |
| **Labels** | Number of distinct label names across all series |
| **Inactive for** | How long since this metric last received a sample |

### Sidebar

The left sidebar shows:
- A **job filter** to narrow the table to metrics from a specific scrape job
- Summary stats: total ghost count, total series, and selected count

### Search and sort

Use the search bar to filter by metric name. Sort by any column using the pills in the top bar — click again to toggle ascending/descending.

### Selecting metrics

Check individual rows or use the header checkbox to select all visible metrics. The banner at the top shows how many series would be dropped if you removed the selected metrics.

---

## Drop Rules

Once you have selected the metrics you want to remove, click **Generate drop rules**. Cardamon will:

1. Group the selected metrics by job
2. Where possible, combine multiple metric names into a single regex to minimise rule count
3. Display each rule with its metric count and series impact

When combining multiple metrics into a single regex, it will make sure that just the specified unused metrics are targeted and not any additional metric with the same prefix, that is used.

### The drop rule modal

The modal shows:
- A list of generated rules, each with a regex, metric count, series count, and job label
- A YAML preview that updates live as you toggle rules on or off
- **Select all / Deselect all** controls
- A **Copy** button and a **Copy & close** button

### Output format

The generated YAML is ready to paste directly into a Prometheus `metric_relabel_configs` block:

```yaml
- source_labels: [__name__]
  regex: "my_unused_metric_total|another_ghost_metric"
  action: drop
```

Add this under the relevant job in your `prometheus.yml` or your scrape config, then reload Prometheus.

---

## Required Permissions

### Prometheus

Cardamon queries the following Prometheus HTTP API endpoints:

- `GET /api/v1/label/__name__/values` Fetch all metric names
- `GET /api/v1/rules` Fetch alerting and recording rules
- `GET /api/v1/series` Fetch series for ghost enrichment
- `GET /api/v1/query` Fetch last-seen timestamps

No write access is required.

### Grafana

Cardamon uses the following Grafana API endpoints:

- `GET /api/search?type=dash-db` List all dashboards
- `GET /api/dashboards/uid/:uid` Fetch dashboard JSON
---

## Contributing

Bug reports, feature requests, and pull requests are welcome. Please open an issue first to discuss significant changes.
