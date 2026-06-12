package core

import (
	"fmt"
	"testing"
)

func testRoute(id RouteID, providerID ProviderID, costPenaltyBps uint16) Route {
	age := uint64(30)
	return Route{
		ID:                 id,
		ProviderID:         providerID,
		Corridor:           NewCorridor("US", "NG"),
		PayoutMethod:       PayoutBankAccount,
		AmountRange:        AmountRange{MinMinorUnits: 1, MaxMinorUnits: 50_000_000},
		SettlementCurrency: "NGN",
		Status:             RouteEnabled,
		LiquidityAvailable: true,
		FXRateAgeSeconds:   &age,
		CostPenaltyBps:     costPenaltyBps,
		FXQualityBps:       9_000,
	}
}

func testInput() EligibilityInput {
	return EligibilityInput{
		TransactionID:               "txn-1",
		Corridor:                    NewCorridor("US", "NG"),
		PayoutMethod:                PayoutBankAccount,
		Amount:                      NewMoney("NGN", 10_000),
		DestinationBank:             "ACCESS",
		CircuitBreakerBlockedRoutes: map[RouteID]bool{},
	}
}

func TestSelectRouteChoosesHighestScoredEligibleRoute(t *testing.T) {
	t.Parallel()

	registry := NewRouteRegistry()
	registry.AddProvider(Provider{ID: "p1", Name: "Provider 1", Enabled: true, Approved: true})
	registry.AddProvider(Provider{ID: "p2", Name: "Provider 2", Enabled: true, Approved: true})
	registry.AddRoute(testRoute("r1", "p1", 2_000))
	registry.AddRoute(testRoute("r2", "p2", 0))

	health := NewRouteHealthBook()
	health.Routes["r1"] = RouteHealthSnapshot{
		SuccessRateBps:        8_200,
		UptimeBps:             10_000,
		P95LatencyMS:          80_000,
		ManualInterventionBps: 0,
		SettlementReliability: 10_000,
	}
	health.Routes["r2"] = RouteHealthSnapshot{
		SuccessRateBps:        9_800,
		UptimeBps:             10_000,
		P95LatencyMS:          30_000,
		ManualInterventionBps: 0,
		SettlementReliability: 10_000,
	}

	decision := SelectRoute(registry, DefaultBankPolicy(), health, testInput())

	if decision.SelectedRoute == nil {
		t.Fatal("expected selected route")
	}
	if decision.SelectedRoute.RouteID != "r2" {
		t.Fatalf("expected r2, got %s", decision.SelectedRoute.RouteID)
	}
	if len(decision.EligibleRoutes) != 2 {
		t.Fatalf("expected 2 eligible routes, got %d", len(decision.EligibleRoutes))
	}
}

func TestSelectRouteRejectsStaleFXAndOpenCircuitBreaker(t *testing.T) {
	t.Parallel()

	registry := NewRouteRegistry()
	registry.AddProvider(Provider{ID: "p1", Name: "Provider 1", Enabled: true, Approved: true})

	stale := testRoute("stale", "p1", 0)
	age := uint64(999)
	stale.FXRateAgeSeconds = &age
	registry.AddRoute(stale)
	registry.AddRoute(testRoute("blocked", "p1", 0))

	input := testInput()
	input.CircuitBreakerBlockedRoutes["blocked"] = true

	decision := SelectRoute(registry, DefaultBankPolicy(), NewRouteHealthBook(), input)

	if decision.SelectedRoute != nil {
		t.Fatalf("expected no selected route, got %#v", decision.SelectedRoute)
	}
	if len(decision.RejectedRoutes) != 2 {
		t.Fatalf("expected 2 rejected routes, got %d", len(decision.RejectedRoutes))
	}

	reasons := map[RejectionReason]bool{}
	for _, route := range decision.RejectedRoutes {
		reasons[route.Reason] = true
	}
	if !reasons[RejectStaleFXRate] {
		t.Fatal("expected stale FX rejection")
	}
	if !reasons[RejectCircuitBreakerOpen] {
		t.Fatal("expected circuit breaker rejection")
	}
}

func BenchmarkSelectRoute100Candidates(b *testing.B) {
	registry := NewRouteRegistry()
	health := NewRouteHealthBook()
	for i := range 100 {
		providerID := ProviderID(fmt.Sprintf("p-%03d", i))
		routeID := RouteID(fmt.Sprintf("r-%03d", i))
		registry.AddProvider(Provider{ID: providerID, Name: string(providerID), Enabled: true, Approved: true})
		registry.AddRoute(testRoute(routeID, providerID, uint16(i*10)))
		health.Routes[routeID] = DefaultRouteHealthSnapshot()
	}

	policy := DefaultBankPolicy()
	input := testInput()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		decision := SelectRoute(registry, policy, health, input)
		if decision.SelectedRoute == nil {
			b.Fatal("expected selected route")
		}
	}
}
