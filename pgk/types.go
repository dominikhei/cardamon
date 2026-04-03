package types

import "time"

type MetricUsage struct {
    Name             string
    Cardinality      int       // From /api/v1/status/tsdb
    InDashboards     []string  // List of Dashboard UIDs
    InAlerts         []string  // List of Alert names
    IsRecordingRule  bool
    LastQueried      time.Time // From query_log
}