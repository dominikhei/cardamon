package audit

type MetricReport struct {
    Name             string        `json:"name"`
    Job              string        `json:"job"`
    SeriesCount      int           `json:"series_count"`
    LabelCount       int           `json:"label_count"`
    InactiveDuration string        `json:"inactive_duration"`
}

