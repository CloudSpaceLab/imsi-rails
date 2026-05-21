package dashboard

import (
	"cmp"
	"math"
	"slices"
	"strings"
	"time"
)

type TransactionMetric struct {
	ID              string
	ProviderID      string
	ProviderName    string
	Corridor        string
	PayoutMethod    string
	Currency        string
	Amount          float64
	ProcessedAt     time.Time
	State           string
	TimeToCreditSec int
	SLALimitSec     int
}

type Service struct {
	transactions []TransactionMetric
	fxRates      map[string]float64
	now          func() time.Time
}

func NewService(transactions []TransactionMetric, fxRates map[string]float64) *Service {
	return &Service{
		transactions: transactions,
		fxRates:      fxRates,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

func NewSeedService() *Service {
	now := time.Now().UTC().Truncate(time.Hour)
	return NewService(seedTransactions(now), map[string]float64{
		"USD": 1,
		"EUR": 1.08,
		"GBP": 1.27,
		"KES": 0.0077,
		"NGN": 0.00064,
	})
}

func (s *Service) Summary(context DashboardContext) SummaryResponse {
	if context.To.IsZero() {
		context.To = s.now()
	}
	if context.From.IsZero() {
		context.From = context.To.Add(-24 * time.Hour)
	}
	if context.Currency == "" {
		context.Currency = "USD"
	}
	records := s.filter(context)
	analytics := s.analytics(records, context.Currency)
	return SummaryResponse{
		Context:     context,
		Analytics:   analytics,
		Tiles:       tilesFor(analytics, context),
		Providers:   s.providerComparisons(records, context.Currency),
		GeneratedAt: s.now(),
	}
}

func (s *Service) Analytics(context DashboardContext) DashboardAnalytics {
	return s.analytics(s.filter(context), context.Currency)
}

func (s *Service) TimeSeries(context DashboardContext) []TimeSeriesPoint {
	if context.To.IsZero() {
		context.To = s.now()
	}
	if context.From.IsZero() {
		context.From = context.To.Add(-24 * time.Hour)
	}
	if context.Currency == "" {
		context.Currency = "USD"
	}
	records := s.filter(context)
	buckets := map[time.Time][]TransactionMetric{}
	for _, record := range records {
		bucket := record.ProcessedAt.Truncate(time.Hour)
		buckets[bucket] = append(buckets[bucket], record)
	}
	points := make([]TimeSeriesPoint, 0)
	for cursor := context.From.Truncate(time.Hour); !cursor.After(context.To); cursor = cursor.Add(time.Hour) {
		analytics := s.analytics(buckets[cursor], context.Currency)
		points = append(points, TimeSeriesPoint{
			Time:           cursor,
			ProcessedCount: analytics.ProcessedCount,
			Volume:         analytics.ProcessedVolume,
			SLARate:        analytics.SLARate,
			P95Seconds:     analytics.P95Seconds,
			State:          stateForAnalytics(analytics),
		})
	}
	return points
}

func (s *Service) filter(context DashboardContext) []TransactionMetric {
	filtered := make([]TransactionMetric, 0)
	for _, record := range s.transactions {
		if !context.From.IsZero() && record.ProcessedAt.Before(context.From) {
			continue
		}
		if !context.To.IsZero() && record.ProcessedAt.After(context.To) {
			continue
		}
		if context.ProviderID != "" && !strings.EqualFold(record.ProviderID, context.ProviderID) {
			continue
		}
		if context.Corridor != "" && !strings.EqualFold(record.Corridor, context.Corridor) {
			continue
		}
		if context.PayoutMethod != "" && !strings.EqualFold(record.PayoutMethod, context.PayoutMethod) {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered
}

func (s *Service) analytics(records []TransactionMetric, currency string) DashboardAnalytics {
	analytics := DashboardAnalytics{ProcessedCount: len(records)}
	latencies := make([]int, 0, len(records))
	for _, record := range records {
		analytics.ProcessedVolume += s.convert(record.Amount, record.Currency, currency)
		switch record.State {
		case "credited", "reconciled", "paid_cash":
			analytics.CompletedCount++
			latencies = append(latencies, record.TimeToCreditSec)
			if record.TimeToCreditSec <= record.SLALimitSec {
				analytics.SLACompletedCount++
			}
		case "failed", "reversed", "disputed":
			analytics.FailedCount++
		case "created", "accepted", "sent_to_provider", "provider_accepted", "payout_submitted":
			analytics.PendingCount++
		default:
			analytics.StalledCount++
		}
	}
	if analytics.CompletedCount > 0 {
		analytics.SLARate = (float64(analytics.SLACompletedCount) / float64(analytics.CompletedCount)) * 100
	}
	slices.Sort(latencies)
	analytics.P50Seconds = percentile(latencies, 0.50)
	analytics.P95Seconds = percentile(latencies, 0.95)
	analytics.P99Seconds = percentile(latencies, 0.99)
	return analytics
}

func (s *Service) providerComparisons(records []TransactionMetric, currency string) []ProviderComparison {
	grouped := map[string][]TransactionMetric{}
	for _, record := range records {
		grouped[record.ProviderID] = append(grouped[record.ProviderID], record)
	}
	comparisons := make([]ProviderComparison, 0, len(grouped))
	for providerID, providerRecords := range grouped {
		analytics := s.analytics(providerRecords, currency)
		name := providerID
		corridor := ""
		if len(providerRecords) > 0 {
			name = providerRecords[0].ProviderName
			corridor = providerRecords[0].Corridor
		}
		comparisons = append(comparisons, ProviderComparison{
			ProviderID:        providerID,
			ProviderName:      name,
			Corridor:          corridor,
			ProcessedCount:    analytics.ProcessedCount,
			ProcessedVolume:   analytics.ProcessedVolume,
			SLACompletedCount: analytics.SLACompletedCount,
			SLARate:           analytics.SLARate,
			P95Seconds:        analytics.P95Seconds,
			State:             stateForAnalytics(analytics),
		})
	}
	slices.SortFunc(comparisons, func(left, right ProviderComparison) int {
		if stateOrder(left.State) != stateOrder(right.State) {
			return cmp.Compare(stateOrder(left.State), stateOrder(right.State))
		}
		return cmp.Compare(right.ProcessedCount, left.ProcessedCount)
	})
	return comparisons
}

func (s *Service) convert(amount float64, from, to string) float64 {
	fromRate := s.fxRates[strings.ToUpper(from)]
	toRate := s.fxRates[strings.ToUpper(to)]
	if fromRate == 0 {
		fromRate = 1
	}
	if toRate == 0 {
		toRate = 1
	}
	return (amount * fromRate) / toRate
}

func tilesFor(analytics DashboardAnalytics, context DashboardContext) []MetricTile {
	currency := context.Currency
	return []MetricTile{
		{ID: "processed", Label: "Processed", Value: formatInt(analytics.ProcessedCount), Unit: "txns", State: HealthHealthy, Trend: "selected range", Drilldown: "/transactions"},
		{ID: "volume", Label: "Volume", Value: formatMoney(analytics.ProcessedVolume), Unit: currency, State: HealthHealthy, Trend: "gross processed value", Drilldown: "/transactions?metric=volume"},
		{ID: "sla", Label: "Completed in SLA", Value: formatPercent(analytics.SLARate), Unit: "%", State: stateForSLA(analytics.SLARate), Trend: formatInt(analytics.SLACompletedCount) + " completed on time", Drilldown: "/transactions?timing=Under+QA+policy"},
		{ID: "p95", Label: "P95 credit time", Value: formatSeconds(analytics.P95Seconds), Unit: "sec", State: stateForP95(analytics.P95Seconds), Trend: "credited transactions", Drilldown: "/incidents?focus=latency"},
		{ID: "failed", Label: "Failed/stalled", Value: formatInt(analytics.FailedCount + analytics.StalledCount), Unit: "txns", State: stateForFailures(analytics), Trend: "needs review", Drilldown: "/transactions?timing=Stalled+only"},
	}
}

func percentile(values []int, p float64) int {
	if len(values) == 0 {
		return 0
	}
	index := int(math.Ceil(float64(len(values))*p)) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

func stateForAnalytics(analytics DashboardAnalytics) HealthState {
	if analytics.FailedCount+analytics.StalledCount > analytics.ProcessedCount/10 && analytics.ProcessedCount > 0 {
		return HealthDegraded
	}
	if analytics.SLARate < 90 && analytics.CompletedCount > 0 {
		return HealthWatch
	}
	return HealthHealthy
}

func stateForSLA(rate float64) HealthState {
	switch {
	case rate >= 95:
		return HealthHealthy
	case rate >= 90:
		return HealthWatch
	default:
		return HealthDegraded
	}
}

func stateForP95(seconds int) HealthState {
	switch {
	case seconds == 0 || seconds <= 90:
		return HealthHealthy
	case seconds <= 180:
		return HealthWatch
	default:
		return HealthDegraded
	}
}

func stateForFailures(analytics DashboardAnalytics) HealthState {
	if analytics.FailedCount+analytics.StalledCount == 0 {
		return HealthHealthy
	}
	if analytics.FailedCount+analytics.StalledCount > analytics.ProcessedCount/10 {
		return HealthDegraded
	}
	return HealthWatch
}

func stateOrder(state HealthState) int {
	switch state {
	case HealthBlocked:
		return 0
	case HealthDegraded:
		return 1
	case HealthWatch:
		return 2
	case HealthRecovery:
		return 3
	default:
		return 4
	}
}
