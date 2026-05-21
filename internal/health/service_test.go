package health

import (
	"errors"
	"testing"
	"time"

	"github.com/CloudSpaceLab/imsi-rails/internal/core"
)

func testHealthService() (*Service, *InMemoryStore, *InMemoryEventSink) {
	store := NewInMemoryStore()
	events := NewInMemoryEventSink()
	return NewService(store, events), store, events
}

func uint16Ptr(value uint16) *uint16 {
	return &value
}

func TestIngestProviderAPIStatusEmitsStateChange(t *testing.T) {
	t.Parallel()

	service, store, events := testHealthService()
	response, err := service.Ingest(IngestSampleRequest{
		ProviderID: "thunes",
		SignalType: SignalProviderAPIStatus,
		ProviderAPI: &ProviderAPISignal{
			Status:    ProviderAPIDown,
			LatencyMS: 5_000,
		},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	if response.Sample.State != StateDegraded {
		t.Fatalf("expected degraded state, got %s", response.Sample.State)
	}
	if response.StateChange == nil {
		t.Fatal("expected state-change event")
	}
	if response.StateChange.PreviousState != StateUnknown {
		t.Fatalf("expected previous unknown, got %s", response.StateChange.PreviousState)
	}
	if got := len(events.Events()); got != 1 {
		t.Fatalf("expected 1 state-change event, got %d", got)
	}
	if got := len(store.Samples()); got != 1 {
		t.Fatalf("expected 1 stored sample, got %d", got)
	}
}

func TestIngestTimeoutRateUpdatesRouteSnapshot(t *testing.T) {
	t.Parallel()

	service, _, events := testHealthService()
	observedAt := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	response, err := service.Ingest(IngestSampleRequest{
		ProviderID:    "ria",
		RouteID:       "eu-ng-account",
		SignalType:    SignalTimeoutRate,
		ObservedAt:    observedAt,
		WindowSeconds: 60,
		TimeoutRate: &RateSignal{
			RateBps: uint16Ptr(1_250),
		},
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	if response.Sample.State != StateDegraded {
		t.Fatalf("expected degraded state, got %s", response.Sample.State)
	}
	if response.Sample.RouteSnapshot == nil {
		t.Fatal("expected route snapshot")
	}
	if response.Sample.RouteSnapshot.SuccessRateBps != 8_750 {
		t.Fatalf("expected success rate 8750, got %d", response.Sample.RouteSnapshot.SuccessRateBps)
	}
	if response.Sample.RouteSnapshot.ObservedAt != observedAt {
		t.Fatalf("expected observed_at %s, got %s", observedAt, response.Sample.RouteSnapshot.ObservedAt)
	}

	route, ok := service.LatestRoute("eu-ng-account")
	if !ok {
		t.Fatal("expected latest route health")
	}
	if route.State != StateDegraded {
		t.Fatalf("expected latest route degraded, got %s", route.State)
	}
	if route.Snapshot.SuccessRateBps != 8_750 {
		t.Fatalf("expected latest success rate 8750, got %d", route.Snapshot.SuccessRateBps)
	}
	if service.SnapshotFor("eu-ng-account").SuccessRateBps != 8_750 {
		t.Fatal("expected routing health provider to expose latest snapshot")
	}
	if got := len(events.Events()); got != 1 {
		t.Fatalf("expected 1 state-change event, got %d", got)
	}
	if got := len(events.CircuitBreakerEvents()); got != 1 {
		t.Fatalf("expected 1 breaker event, got %d", got)
	}
}

func TestIngestCallbackLagAndTransactionOutcome(t *testing.T) {
	t.Parallel()

	service, _, events := testHealthService()
	_, err := service.Ingest(IngestSampleRequest{
		ProviderID:  "remitly",
		RouteID:     "uk-ng-account",
		SignalType:  SignalCallbackLag,
		CallbackLag: &CallbackLagSignal{P95LagMS: 75_000},
	})
	if err != nil {
		t.Fatalf("callback lag ingest failed: %v", err)
	}

	response, err := service.Ingest(IngestSampleRequest{
		ProviderID: "remitly",
		RouteID:    "uk-ng-account",
		SignalType: SignalTransactionOutcome,
		TransactionOutcome: &TransactionOutcome{
			TransactionID:  "txn_1",
			State:          core.StateCredited,
			TimeToCreditMS: 40_000,
		},
	})
	if err != nil {
		t.Fatalf("transaction outcome ingest failed: %v", err)
	}
	if response.Sample.State != StateHealthy {
		t.Fatalf("expected healthy state, got %s", response.Sample.State)
	}

	route, ok := service.LatestRoute("uk-ng-account")
	if !ok {
		t.Fatal("expected latest route health")
	}
	if route.State != StateHealthy {
		t.Fatalf("expected route state healthy, got %s", route.State)
	}
	if route.Snapshot.P95LatencyMS != 40_000 {
		t.Fatalf("expected p95 latency 40000, got %d", route.Snapshot.P95LatencyMS)
	}
	if got := len(events.Events()); got != 2 {
		t.Fatalf("expected 2 state-change events, got %d", got)
	}
	if got := len(events.CircuitBreakerEvents()); got != 2 {
		t.Fatalf("expected 2 breaker events, got %d", got)
	}
}

func TestIngestSameStateDoesNotEmitDuplicateChange(t *testing.T) {
	t.Parallel()

	service, _, events := testHealthService()
	request := IngestSampleRequest{
		ProviderID: "thunes",
		RouteID:    "us-ng-account",
		SignalType: SignalProviderAPIStatus,
		ProviderAPI: &ProviderAPISignal{
			Status:    ProviderAPIUp,
			LatencyMS: 900,
		},
	}
	if _, err := service.Ingest(request); err != nil {
		t.Fatalf("first ingest failed: %v", err)
	}
	response, err := service.Ingest(request)
	if err != nil {
		t.Fatalf("second ingest failed: %v", err)
	}
	if response.StateChange != nil {
		t.Fatal("did not expect duplicate state-change event")
	}
	if got := len(events.Events()); got != 1 {
		t.Fatalf("expected 1 state-change event, got %d", got)
	}
}

func TestIngestRejectsMismatchedPayload(t *testing.T) {
	t.Parallel()

	service, _, _ := testHealthService()
	_, err := service.Ingest(IngestSampleRequest{
		ProviderID:  "thunes",
		SignalType:  SignalTimeoutRate,
		CallbackLag: &CallbackLagSignal{P95LagMS: 1_000},
	})
	if !errors.Is(err, ErrInvalidSignalPayload) {
		t.Fatalf("expected invalid payload error, got %v", err)
	}
}
