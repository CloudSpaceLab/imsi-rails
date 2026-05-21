package core

import "testing"

func TestMoneyAndCorridorNormalizeCodes(t *testing.T) {
	t.Parallel()

	money := NewMoney("usd", 10_000)
	corridor := NewCorridor("us", "ng")

	if money.Currency != "USD" {
		t.Fatalf("expected USD, got %s", money.Currency)
	}
	if corridor.OriginCountry != "US" || corridor.DestinationCountry != "NG" {
		t.Fatalf("unexpected corridor: %#v", corridor)
	}
}

func TestRouteDestinationBankSupport(t *testing.T) {
	t.Parallel()

	age := uint64(10)
	route := Route{
		ID:                 "route-1",
		ProviderID:         "provider-1",
		Corridor:           NewCorridor("US", "NG"),
		PayoutMethod:       PayoutBankAccount,
		AmountRange:        AmountRange{MinMinorUnits: 1, MaxMinorUnits: 1_000_000},
		SettlementCurrency: "NGN",
		DestinationBanks:   []string{"ZENITH"},
		Status:             RouteEnabled,
		LiquidityAvailable: true,
		FXRateAgeSeconds:   &age,
		FXQualityBps:       9_000,
	}

	if !route.SupportsDestinationBank("zenith") {
		t.Fatal("route should support zenith")
	}
	if route.SupportsDestinationBank("gtbank") {
		t.Fatal("route should not support gtbank")
	}
	if route.SupportsDestinationBank("") {
		t.Fatal("route should not support an empty destination bank when limited")
	}
}
