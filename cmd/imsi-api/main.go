package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/CloudSpaceLab/imsi-rails/internal/auth"
	"github.com/CloudSpaceLab/imsi-rails/internal/core"
	"github.com/CloudSpaceLab/imsi-rails/internal/dashboard"
	"github.com/CloudSpaceLab/imsi-rails/internal/health"
	"github.com/CloudSpaceLab/imsi-rails/internal/intake"
)

func main() {
	mux := http.NewServeMux()
	authService, err := auth.NewDefaultService()
	if err != nil {
		log.Fatal(err)
	}
	healthService := seedHealthService()
	auth.NewHandler(authService).Register(mux)
	intake.NewHandler(seedIntakeService(healthService)).RegisterProtected(mux, authService.Require)
	health.NewHandler(healthService).RegisterProtected(mux, authService.Require)
	dashboard.NewHandler(dashboard.NewSeedService()).RegisterProtected(mux, authService.Require)

	addr := ":8080"
	if envAddr := os.Getenv("IMSI_API_ADDR"); envAddr != "" {
		addr = envAddr
	}

	log.Printf("imsi-rails API listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func seedIntakeService(healthSource core.RouteHealthProvider) *intake.Service {
	registry := core.NewRouteRegistry()
	registry.AddProvider(core.Provider{ID: "thunes", Name: "Thunes", Enabled: true, Approved: true})
	registry.AddProvider(core.Provider{ID: "remitly", Name: "Remitly", Enabled: true, Approved: true})
	registry.AddProvider(core.Provider{ID: "ria", Name: "Ria", Enabled: true, Approved: true})
	registry.AddProvider(core.Provider{ID: "papss", Name: "PAPSS", Enabled: true, Approved: true})
	age := uint64(30)
	registry.AddRoute(core.Route{
		ID:                 "thunes-us-ng-account",
		ProviderID:         "thunes",
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
	registry.AddRoute(core.Route{
		ID:                 "remitly-gb-ng-account",
		ProviderID:         "remitly",
		Corridor:           core.NewCorridor("GB", "NG"),
		PayoutMethod:       core.PayoutBankAccount,
		AmountRange:        core.AmountRange{MinMinorUnits: 1, MaxMinorUnits: 30_000_000},
		SettlementCurrency: "NGN",
		Status:             core.RouteEnabled,
		LiquidityAvailable: true,
		FXRateAgeSeconds:   &age,
		CostPenaltyBps:     90,
		FXQualityBps:       8_900,
	})
	registry.AddRoute(core.Route{
		ID:                 "ria-eu-ng-account",
		ProviderID:         "ria",
		Corridor:           core.NewCorridor("EU", "NG"),
		PayoutMethod:       core.PayoutBankAccount,
		AmountRange:        core.AmountRange{MinMinorUnits: 1, MaxMinorUnits: 40_000_000},
		SettlementCurrency: "NGN",
		Status:             core.RouteEnabled,
		LiquidityAvailable: true,
		FXRateAgeSeconds:   &age,
		CostPenaltyBps:     74,
		FXQualityBps:       8_600,
	})
	registry.AddRoute(core.Route{
		ID:                 "papss-ke-ng-account",
		ProviderID:         "papss",
		Corridor:           core.NewCorridor("KE", "NG"),
		PayoutMethod:       core.PayoutBankAccount,
		AmountRange:        core.AmountRange{MinMinorUnits: 1, MaxMinorUnits: 80_000_000},
		SettlementCurrency: "NGN",
		Status:             core.RouteRecoveryTesting,
		LiquidityAvailable: true,
		FXRateAgeSeconds:   &age,
		CostPenaltyBps:     68,
		FXQualityBps:       8_800,
	})

	return intake.NewService(
		registry,
		core.DefaultBankPolicy(),
		healthSource,
		intake.NewInMemoryTransactionStore(),
		core.NewInMemoryRouteDecisionStore(),
		intake.NewInMemoryEventSink(),
	)
}

func seedHealthService() *health.Service {
	return health.NewService(
		health.NewInMemoryStore(),
		health.NewInMemoryEventSink(),
	)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	if origin == "https://imsi.cloudspacetechs.com" {
		return true
	}
	if os.Getenv("IMSI_ENV") != "production" {
		return strings.HasPrefix(origin, "http://127.0.0.1:") || strings.HasPrefix(origin, "http://localhost:")
	}
	return false
}
