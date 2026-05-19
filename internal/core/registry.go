package core

import "strings"

type ProviderID string
type RouteID string

type Money struct {
	Currency   string `json:"currency"`
	MinorUnits uint64 `json:"minor_units"`
}

func NewMoney(currency string, minorUnits uint64) Money {
	return Money{
		Currency:   strings.ToUpper(currency),
		MinorUnits: minorUnits,
	}
}

type AmountRange struct {
	MinMinorUnits uint64 `json:"min_minor_units"`
	MaxMinorUnits uint64 `json:"max_minor_units"`
}

func (r AmountRange) Contains(amount Money) bool {
	return amount.MinorUnits >= r.MinMinorUnits && amount.MinorUnits <= r.MaxMinorUnits
}

type Corridor struct {
	OriginCountry      string `json:"origin_country"`
	DestinationCountry string `json:"destination_country"`
}

func NewCorridor(originCountry, destinationCountry string) Corridor {
	return Corridor{
		OriginCountry:      strings.ToUpper(originCountry),
		DestinationCountry: strings.ToUpper(destinationCountry),
	}
}

type PayoutMethod string

const (
	PayoutBankAccount PayoutMethod = "bank_account"
	PayoutCashPickup  PayoutMethod = "cash_pickup"
	PayoutWallet      PayoutMethod = "wallet"
	PayoutCard        PayoutMethod = "card"
	PayoutBranch      PayoutMethod = "branch"
)

type RouteStatus string

const (
	RouteEnabled         RouteStatus = "enabled"
	RouteDisabled        RouteStatus = "disabled"
	RouteDegraded        RouteStatus = "degraded"
	RouteBlocked         RouteStatus = "blocked"
	RouteRecoveryTesting RouteStatus = "recovery_testing"
)

type Provider struct {
	ID       ProviderID `json:"id"`
	Name     string     `json:"name"`
	Enabled  bool       `json:"enabled"`
	Approved bool       `json:"approved"`
}

type Route struct {
	ID                 RouteID      `json:"id"`
	ProviderID         ProviderID   `json:"provider_id"`
	Corridor           Corridor     `json:"corridor"`
	PayoutMethod       PayoutMethod `json:"payout_method"`
	AmountRange        AmountRange  `json:"amount_range"`
	SettlementCurrency string       `json:"settlement_currency"`
	DestinationBanks   []string     `json:"destination_banks"`
	Status             RouteStatus  `json:"status"`
	LiquidityAvailable bool         `json:"liquidity_available"`
	FXRateAgeSeconds   *uint64      `json:"fx_rate_age_seconds,omitempty"`
	CostPenaltyBps     uint16       `json:"cost_penalty_bps"`
	FXQualityBps       uint16       `json:"fx_quality_bps"`
}

func (r Route) SupportsDestinationBank(destinationBank string) bool {
	if len(r.DestinationBanks) == 0 {
		return true
	}

	if destinationBank == "" {
		return false
	}

	normalized := strings.ToUpper(destinationBank)
	for _, bank := range r.DestinationBanks {
		if bank == normalized {
			return true
		}
	}
	return false
}

type RouteRegistry struct {
	Providers map[ProviderID]Provider `json:"providers"`
	Routes    map[RouteID]Route       `json:"routes"`
}

func NewRouteRegistry() RouteRegistry {
	return RouteRegistry{
		Providers: map[ProviderID]Provider{},
		Routes:    map[RouteID]Route{},
	}
}

func (r *RouteRegistry) AddProvider(provider Provider) {
	r.Providers[provider.ID] = provider
}

func (r *RouteRegistry) AddRoute(route Route) {
	r.Routes[route.ID] = route
}

func (r RouteRegistry) Provider(id ProviderID) (Provider, bool) {
	provider, ok := r.Providers[id]
	return provider, ok
}

func (r RouteRegistry) RoutesForCorridor(corridor Corridor) []Route {
	routes := make([]Route, 0)
	for _, route := range r.Routes {
		if route.Corridor == corridor {
			routes = append(routes, route)
		}
	}
	return routes
}
