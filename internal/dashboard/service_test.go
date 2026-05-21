package dashboard

import (
	"testing"
	"time"
)

func TestDashboardAnalyticsForSelectedRangeAndProvider(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	service := NewService([]TransactionMetric{
		{ID: "1", ProviderID: "thunes", ProviderName: "Thunes", Corridor: "US -> NG", PayoutMethod: "bank_account", Currency: "USD", Amount: 100, ProcessedAt: now.Add(-time.Hour), State: "credited", TimeToCreditSec: 40, SLALimitSec: 90},
		{ID: "2", ProviderID: "thunes", ProviderName: "Thunes", Corridor: "US -> NG", PayoutMethod: "bank_account", Currency: "USD", Amount: 200, ProcessedAt: now.Add(-30 * time.Minute), State: "credited", TimeToCreditSec: 120, SLALimitSec: 90},
		{ID: "3", ProviderID: "ria", ProviderName: "Ria", Corridor: "EU -> NG", PayoutMethod: "bank_account", Currency: "EUR", Amount: 300, ProcessedAt: now.Add(-20 * time.Minute), State: "failed", SLALimitSec: 90},
	}, map[string]float64{"USD": 1, "EUR": 1.1})

	summary := service.Summary(DashboardContext{From: now.Add(-2 * time.Hour), To: now, ProviderID: "thunes", Currency: "USD"})
	if summary.Analytics.ProcessedCount != 2 {
		t.Fatalf("processed count = %d", summary.Analytics.ProcessedCount)
	}
	if summary.Analytics.SLACompletedCount != 1 {
		t.Fatalf("sla completed = %d", summary.Analytics.SLACompletedCount)
	}
	if summary.Analytics.SLARate != 50 {
		t.Fatalf("sla rate = %f", summary.Analytics.SLARate)
	}
	if summary.Analytics.ProcessedVolume != 300 {
		t.Fatalf("volume = %f", summary.Analytics.ProcessedVolume)
	}
	if len(summary.Tiles) == 0 || summary.Tiles[0].Drilldown == "" {
		t.Fatalf("expected clickable metric tiles")
	}
}

func TestDashboardTimeseriesBuckets(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	service := NewService([]TransactionMetric{
		{ID: "1", ProviderID: "thunes", ProviderName: "Thunes", Corridor: "US -> NG", PayoutMethod: "bank_account", Currency: "USD", Amount: 100, ProcessedAt: now.Add(-90 * time.Minute), State: "credited", TimeToCreditSec: 40, SLALimitSec: 90},
		{ID: "2", ProviderID: "thunes", ProviderName: "Thunes", Corridor: "US -> NG", PayoutMethod: "bank_account", Currency: "USD", Amount: 100, ProcessedAt: now.Add(-30 * time.Minute), State: "credited", TimeToCreditSec: 80, SLALimitSec: 90},
	}, map[string]float64{"USD": 1})

	points := service.TimeSeries(DashboardContext{From: now.Add(-2 * time.Hour), To: now, Currency: "USD"})
	if len(points) != 3 {
		t.Fatalf("expected three hourly buckets, got %d", len(points))
	}
	if points[1].ProcessedCount == 0 && points[2].ProcessedCount == 0 {
		t.Fatalf("expected populated buckets: %#v", points)
	}
}
