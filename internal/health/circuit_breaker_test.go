package health

import (
	"testing"
	"time"

	"github.com/CloudSpaceLab/imsi-rails/internal/core"
)

func TestCircuitBreakerBlocksDegradedSnapshot(t *testing.T) {
	t.Parallel()

	observedAt := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	record := EvaluateCircuitBreaker(
		"ria-eu-ng",
		CircuitBreakerRecord{State: BreakerUnknown},
		core.RouteHealthSnapshot{
			SuccessRateBps:        7_500,
			UptimeBps:             9_900,
			P95LatencyMS:          2_000,
			ManualInterventionBps: 0,
			SettlementReliability: 10_000,
			ObservedAt:            observedAt,
		},
		DefaultCircuitBreakerThresholds(),
		observedAt,
	)

	if record.State != BreakerBlocked {
		t.Fatalf("expected blocked, got %s", record.State)
	}
	if record.Reason != "success_rate_below_blocked_threshold" {
		t.Fatalf("unexpected reason %s", record.Reason)
	}
}

func TestCircuitBreakerMovesBlockedRouteThroughRecoveryTesting(t *testing.T) {
	t.Parallel()

	observedAt := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	thresholds := DefaultCircuitBreakerThresholds()
	recoveredSnapshot := core.RouteHealthSnapshot{
		SuccessRateBps:        9_900,
		UptimeBps:             10_000,
		P95LatencyMS:          30_000,
		ManualInterventionBps: 0,
		SettlementReliability: 10_000,
		ObservedAt:            observedAt,
	}

	recovery := EvaluateCircuitBreaker(
		"ria-eu-ng",
		CircuitBreakerRecord{RouteID: "ria-eu-ng", State: BreakerBlocked},
		recoveredSnapshot,
		thresholds,
		observedAt,
	)
	if recovery.State != BreakerRecoveryTesting {
		t.Fatalf("expected recovery testing, got %s", recovery.State)
	}

	healthy := EvaluateCircuitBreaker(
		"ria-eu-ng",
		recovery,
		recoveredSnapshot,
		thresholds,
		observedAt.Add(time.Minute),
	)
	if healthy.State != BreakerHealthy {
		t.Fatalf("expected healthy after recovery pass, got %s", healthy.State)
	}
}

func TestCircuitBreakerThresholdsAreConfigurable(t *testing.T) {
	t.Parallel()

	observedAt := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	thresholds := DefaultCircuitBreakerThresholds()
	thresholds.BlockedP95LatencyMS = 80_000
	record := EvaluateCircuitBreaker(
		"remitly-uk-ng",
		CircuitBreakerRecord{State: BreakerUnknown},
		core.RouteHealthSnapshot{
			SuccessRateBps:        9_900,
			UptimeBps:             10_000,
			P95LatencyMS:          85_000,
			ManualInterventionBps: 0,
			SettlementReliability: 10_000,
			ObservedAt:            observedAt,
		},
		thresholds,
		observedAt,
	)

	if record.State != BreakerBlocked {
		t.Fatalf("expected configurable latency threshold to block route, got %s", record.State)
	}

	service := NewServiceWithCircuitBreakerThresholds(NewInMemoryStore(), NewInMemoryEventSink(), thresholds)
	response, err := service.Ingest(IngestSampleRequest{
		ProviderID:  "remitly",
		RouteID:     "remitly-uk-ng",
		SignalType:  SignalCallbackLag,
		CallbackLag: &CallbackLagSignal{P95LagMS: 85_000},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if response.CircuitBreaker == nil || response.CircuitBreaker.State != BreakerBlocked {
		t.Fatalf("expected service threshold to block route, got %#v", response.CircuitBreaker)
	}
}
