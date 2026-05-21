package dashboard

import "time"

func seedTransactions(now time.Time) []TransactionMetric {
	base := now.Add(-18 * time.Hour)
	providers := []struct {
		id       string
		name     string
		corridor string
		currency string
		amount   float64
		latency  int
		sla      int
		state    string
	}{
		{"thunes", "Thunes", "US -> NG", "USD", 950, 31, 90, "credited"},
		{"thunes", "Thunes", "US -> NG", "USD", 1210, 42, 90, "credited"},
		{"remitly", "Remitly", "UK -> NG", "GBP", 850, 49, 90, "credited"},
		{"remitly", "Remitly", "UK -> NG", "GBP", 620, 112, 90, "credited"},
		{"ria", "Ria", "EU -> NG", "EUR", 2400, 258, 90, "credited"},
		{"ria", "Ria", "EU -> NG", "EUR", 1800, 464, 90, "credited"},
		{"ria", "Ria", "EU -> NG", "EUR", 730, 0, 90, "provider_accepted"},
		{"papss", "PAPSS", "KE -> NG", "KES", 180000, 56, 90, "credited"},
		{"papss", "PAPSS", "KE -> NG", "KES", 96000, 0, 90, "failed"},
	}

	records := make([]TransactionMetric, 0, 72)
	for hour := 0; hour < 24; hour++ {
		for index, provider := range providers {
			if (hour+index)%4 == 0 && provider.id == "papss" {
				continue
			}
			state := provider.state
			latency := provider.latency + ((hour + index) % 5 * 4)
			if provider.id == "ria" && hour > 12 {
				latency += 80
			}
			records = append(records, TransactionMetric{
				ID:              "txn_seed_" + string(rune('a'+index)) + "_" + time.Duration(hour).String(),
				ProviderID:      provider.id,
				ProviderName:    provider.name,
				Corridor:        provider.corridor,
				PayoutMethod:    "bank_account",
				Currency:        provider.currency,
				Amount:          provider.amount + float64(hour*17),
				ProcessedAt:     base.Add(time.Duration(hour) * time.Hour).Add(time.Duration(index*3) * time.Minute),
				State:           state,
				TimeToCreditSec: latency,
				SLALimitSec:     provider.sla,
			})
		}
	}
	return records
}
