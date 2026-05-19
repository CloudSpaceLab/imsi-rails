package main

import (
	"log"
	"net/http"
	"os"

	"github.com/CloudSpaceLab/imsi-rails/internal/core"
	"github.com/CloudSpaceLab/imsi-rails/internal/intake"
)

func main() {
	mux := http.NewServeMux()
	intake.NewHandler(seedIntakeService()).Register(mux)

	addr := ":8080"
	if envAddr := os.Getenv("IMSI_API_ADDR"); envAddr != "" {
		addr = envAddr
	}

	log.Printf("imsi-rails API listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func seedIntakeService() *intake.Service {
	registry := core.NewRouteRegistry()
	registry.AddProvider(core.Provider{ID: "sandbox-provider", Name: "Sandbox Provider", Enabled: true, Approved: true})
	age := uint64(30)
	registry.AddRoute(core.Route{
		ID:                 "sandbox-us-ng-account",
		ProviderID:         "sandbox-provider",
		Corridor:           core.NewCorridor("US", "NG"),
		PayoutMethod:       core.PayoutBankAccount,
		AmountRange:        core.AmountRange{MinMinorUnits: 1, MaxMinorUnits: 50_000_000},
		SettlementCurrency: "NGN",
		Status:             core.RouteEnabled,
		LiquidityAvailable: true,
		FXRateAgeSeconds:   &age,
		CostPenaltyBps:     0,
		FXQualityBps:       9_200,
	})

	return intake.NewService(
		registry,
		core.DefaultBankPolicy(),
		core.NewRouteHealthBook(),
		intake.NewInMemoryTransactionStore(),
		core.NewInMemoryRouteDecisionStore(),
		intake.NewInMemoryEventSink(),
	)
}
