package dashboard

import "time"

type HealthState string

const (
	HealthHealthy  HealthState = "healthy"
	HealthWatch    HealthState = "watch"
	HealthDegraded HealthState = "degraded"
	HealthBlocked  HealthState = "blocked"
	HealthRecovery HealthState = "recovery"
)

type DashboardContext struct {
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	Timezone     string    `json:"timezone"`
	ProviderID   string    `json:"provider_id,omitempty"`
	Corridor     string    `json:"corridor,omitempty"`
	PayoutMethod string    `json:"payout_method,omitempty"`
	Currency     string    `json:"currency"`
	AnalysisLens string    `json:"analysis_lens"`
}

type DashboardAnalytics struct {
	ProcessedCount    int     `json:"processed_count"`
	ProcessedVolume   float64 `json:"processed_volume"`
	CompletedCount    int     `json:"completed_count"`
	SLACompletedCount int     `json:"sla_completed_count"`
	SLARate           float64 `json:"sla_rate"`
	FailedCount       int     `json:"failed_count"`
	StalledCount      int     `json:"stalled_count"`
	PendingCount      int     `json:"pending_count"`
	P50Seconds        int     `json:"p50_seconds"`
	P95Seconds        int     `json:"p95_seconds"`
	P99Seconds        int     `json:"p99_seconds"`
}

type MetricTile struct {
	ID        string      `json:"id"`
	Label     string      `json:"label"`
	Value     string      `json:"value"`
	Unit      string      `json:"unit"`
	State     HealthState `json:"state"`
	Trend     string      `json:"trend"`
	Drilldown string      `json:"drilldown"`
}

type TimeSeriesPoint struct {
	Time           time.Time   `json:"time"`
	ProcessedCount int         `json:"processed_count"`
	Volume         float64     `json:"volume"`
	SLARate        float64     `json:"sla_rate"`
	P95Seconds     int         `json:"p95_seconds"`
	State          HealthState `json:"state"`
}

type ProviderComparison struct {
	ProviderID        string      `json:"provider_id"`
	ProviderName      string      `json:"provider_name"`
	Corridor          string      `json:"corridor"`
	ProcessedCount    int         `json:"processed_count"`
	ProcessedVolume   float64     `json:"processed_volume"`
	SLACompletedCount int         `json:"sla_completed_count"`
	SLARate           float64     `json:"sla_rate"`
	P95Seconds        int         `json:"p95_seconds"`
	State             HealthState `json:"state"`
}

type SummaryResponse struct {
	Context     DashboardContext     `json:"context"`
	Analytics   DashboardAnalytics   `json:"analytics"`
	Tiles       []MetricTile         `json:"tiles"`
	Providers   []ProviderComparison `json:"providers"`
	GeneratedAt time.Time            `json:"generated_at"`
}
