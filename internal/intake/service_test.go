package intake

import (
	"errors"
	"testing"

	"github.com/CloudSpaceLab/imsi-rails/internal/core"
	"github.com/CloudSpaceLab/imsi-rails/internal/health"
)

func testService() (*Service, *InMemoryTransactionStore, *InMemoryEventSink, *core.InMemoryRouteDecisionStore) {
	transactions := NewInMemoryTransactionStore()
	events := NewInMemoryEventSink()
	decisions := core.NewInMemoryRouteDecisionStore()
	service := NewService(testRegistry(), core.DefaultBankPolicy(), core.NewRouteHealthBook(), transactions, decisions, events)
	return service, transactions, events, decisions
}

func testRegistry() core.RouteRegistry {
	registry := core.NewRouteRegistry()
	registry.AddProvider(core.Provider{ID: "p1", Name: "Provider 1", Enabled: true, Approved: true})
	age := uint64(30)
	registry.AddRoute(core.Route{
		ID:                 "r1",
		ProviderID:         "p1",
		Corridor:           core.NewCorridor("US", "NG"),
		PayoutMethod:       core.PayoutBankAccount,
		AmountRange:        core.AmountRange{MinMinorUnits: 1, MaxMinorUnits: 50_000_000},
		SettlementCurrency: "NGN",
		Status:             core.RouteEnabled,
		LiquidityAvailable: true,
		FXRateAgeSeconds:   &age,
		FXQualityBps:       9_200,
	})
	return registry
}

func validRequest() SubmitTransactionRequest {
	return SubmitTransactionRequest{
		BankID:           "bank-1",
		IDempotencyKey:   "idem-1",
		SenderCountry:    "US",
		ReceiverCountry:  "NG",
		AmountMinorUnits: 10_000,
		Currency:         "NGN",
		PayoutMethod:     core.PayoutBankAccount,
		DestinationBank:  "ACCESS",
	}
}

func TestSubmitPersistsTransactionDecisionAndEvents(t *testing.T) {
	t.Parallel()

	service, transactions, events, decisions := testService()

	response, err := service.Submit(validRequest())
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	if response.TransactionID == "" {
		t.Fatal("expected transaction id")
	}
	if response.SelectedRouteID != "r1" {
		t.Fatalf("expected selected route r1, got %s", response.SelectedRouteID)
	}
	if response.State != core.StateAccepted {
		t.Fatalf("expected accepted state, got %s", response.State)
	}

	if _, ok := transactions.GetByTransactionID(response.TransactionID); !ok {
		t.Fatal("expected transaction persistence")
	}
	if _, ok := decisions.GetByTransactionID(response.TransactionID); !ok {
		t.Fatal("expected route decision persistence")
	}
	if got := len(events.Events()); got != 2 {
		t.Fatalf("expected 2 lifecycle events, got %d", got)
	}
}

func TestSubmitRequiresIDempotencyKey(t *testing.T) {
	t.Parallel()

	service, _, _, _ := testService()
	request := validRequest()
	request.IDempotencyKey = ""

	_, err := service.Submit(request)
	if !errors.Is(err, ErrMissingIDempotencyKey) {
		t.Fatalf("expected missing idempotency error, got %v", err)
	}
}

func TestDuplicateIDempotencyKeyDoesNotDuplicateProcessing(t *testing.T) {
	t.Parallel()

	service, transactions, events, _ := testService()

	first, err := service.Submit(validRequest())
	if err != nil {
		t.Fatalf("first submit failed: %v", err)
	}
	second, err := service.Submit(validRequest())
	if err != nil {
		t.Fatalf("second submit failed: %v", err)
	}

	if !second.IDempotentReplay {
		t.Fatal("expected idempotent replay")
	}
	if first.TransactionID != second.TransactionID {
		t.Fatalf("expected same transaction id, got %s and %s", first.TransactionID, second.TransactionID)
	}
	if got := len(events.Events()); got != 2 {
		t.Fatalf("expected no duplicate event emission, got %d events", got)
	}
	if _, ok := transactions.GetByIDempotencyKey("bank-1", "idem-1"); !ok {
		t.Fatal("expected idempotency lookup")
	}
}

func TestSubmitRejectsCircuitBreakerBlockedRoute(t *testing.T) {
	t.Parallel()

	healthService := health.NewService(health.NewInMemoryStore(), health.NewInMemoryEventSink())
	if _, err := healthService.Ingest(health.IngestSampleRequest{
		ProviderID: "p1",
		RouteID:    "r1",
		SignalType: health.SignalProviderAPIStatus,
		ProviderAPI: &health.ProviderAPISignal{
			Status:    health.ProviderAPIDown,
			LatencyMS: 5_000,
		},
	}); err != nil {
		t.Fatalf("health ingest failed: %v", err)
	}

	service := NewService(
		testRegistry(),
		core.DefaultBankPolicy(),
		healthService,
		NewInMemoryTransactionStore(),
		core.NewInMemoryRouteDecisionStore(),
		NewInMemoryEventSink(),
	)
	response, err := service.Submit(validRequest())
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	if response.State != core.StateFailed {
		t.Fatalf("expected failed state, got %s", response.State)
	}
	if response.RouteDecision.SelectedRoute != nil {
		t.Fatal("expected no selected route")
	}
	if len(response.RouteDecision.RejectedRoutes) != 1 {
		t.Fatalf("expected 1 rejected route, got %d", len(response.RouteDecision.RejectedRoutes))
	}
	if response.RouteDecision.RejectedRoutes[0].Reason != core.RejectCircuitBreakerOpen {
		t.Fatalf("expected circuit breaker rejection, got %s", response.RouteDecision.RejectedRoutes[0].Reason)
	}
}
