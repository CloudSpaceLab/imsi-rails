package intake

import (
	"errors"
	"testing"

	"github.com/CloudSpaceLab/imsi-rails/internal/core"
)

func testService() (*Service, *InMemoryTransactionStore, *InMemoryEventSink, *core.InMemoryRouteDecisionStore) {
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

	transactions := NewInMemoryTransactionStore()
	events := NewInMemoryEventSink()
	decisions := core.NewInMemoryRouteDecisionStore()
	service := NewService(registry, core.DefaultBankPolicy(), core.NewRouteHealthBook(), transactions, decisions, events)
	return service, transactions, events, decisions
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
