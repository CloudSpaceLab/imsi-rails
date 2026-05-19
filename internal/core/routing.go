package core

import (
	"cmp"
	"slices"
	"time"
)

type ScoringWeights struct {
	Reliability uint16 `json:"reliability"`
	Speed       uint16 `json:"speed"`
	Cost        uint16 `json:"cost"`
	FX          uint16 `json:"fx"`
	Liquidity   uint16 `json:"liquidity"`
	Operations  uint16 `json:"operations"`
}

func DefaultScoringWeights() ScoringWeights {
	return ScoringWeights{
		Reliability: 35,
		Speed:       20,
		Cost:        15,
		FX:          10,
		Liquidity:   10,
		Operations:  10,
	}
}

func (w ScoringWeights) Total() uint32 {
	return uint32(w.Reliability) +
		uint32(w.Speed) +
		uint32(w.Cost) +
		uint32(w.FX) +
		uint32(w.Liquidity) +
		uint32(w.Operations)
}

type BankPolicy struct {
	PolicyID             string              `json:"policy_id"`
	Version              uint32              `json:"version"`
	Weights              ScoringWeights      `json:"weights"`
	DisabledProviders    map[ProviderID]bool `json:"disabled_providers"`
	DisabledRoutes       map[RouteID]bool    `json:"disabled_routes"`
	MaxFXAgeSeconds      *uint64             `json:"max_fx_age_seconds,omitempty"`
	MinSuccessRateBps    *uint16             `json:"min_success_rate_bps,omitempty"`
	TargetP95LatencyMS   uint64              `json:"target_p95_latency_ms"`
	AutoSwitchingAllowed bool                `json:"auto_switching_allowed"`
}

func DefaultBankPolicy() BankPolicy {
	maxFXAge := uint64(180)
	minSuccess := uint16(8000)
	return BankPolicy{
		PolicyID:             "default",
		Version:              1,
		Weights:              DefaultScoringWeights(),
		DisabledProviders:    map[ProviderID]bool{},
		DisabledRoutes:       map[RouteID]bool{},
		MaxFXAgeSeconds:      &maxFXAge,
		MinSuccessRateBps:    &minSuccess,
		TargetP95LatencyMS:   120_000,
		AutoSwitchingAllowed: true,
	}
}

type EligibilityInput struct {
	TransactionID                  string           `json:"transaction_id"`
	Corridor                       Corridor         `json:"corridor"`
	PayoutMethod                   PayoutMethod     `json:"payout_method"`
	Amount                         Money            `json:"amount"`
	DestinationBank                string           `json:"destination_bank,omitempty"`
	ComplianceManualReviewRequired bool             `json:"compliance_manual_review_required"`
	CircuitBreakerBlockedRoutes    map[RouteID]bool `json:"circuit_breaker_blocked_routes"`
}

type RouteHealthSnapshot struct {
	SuccessRateBps        uint16    `json:"success_rate_bps"`
	UptimeBps             uint16    `json:"uptime_bps"`
	P95LatencyMS          uint64    `json:"p95_latency_ms"`
	ManualInterventionBps uint16    `json:"manual_intervention_bps"`
	SettlementReliability uint16    `json:"settlement_reliability_bps"`
	ObservedAt            time.Time `json:"observed_at"`
}

func DefaultRouteHealthSnapshot() RouteHealthSnapshot {
	return RouteHealthSnapshot{
		SuccessRateBps:        10_000,
		UptimeBps:             10_000,
		P95LatencyMS:          1_000,
		ManualInterventionBps: 0,
		SettlementReliability: 10_000,
		ObservedAt:            time.Now().UTC(),
	}
}

type RouteHealthBook struct {
	Routes map[RouteID]RouteHealthSnapshot `json:"routes"`
}

type RouteHealthProvider interface {
	SnapshotFor(routeID RouteID) RouteHealthSnapshot
}

func NewRouteHealthBook() RouteHealthBook {
	return RouteHealthBook{Routes: map[RouteID]RouteHealthSnapshot{}}
}

func (b RouteHealthBook) SnapshotFor(routeID RouteID) RouteHealthSnapshot {
	if snapshot, ok := b.Routes[routeID]; ok {
		return snapshot
	}
	return DefaultRouteHealthSnapshot()
}

type RejectionReason string

const (
	RejectProviderMissing        RejectionReason = "provider_missing"
	RejectProviderDisabled       RejectionReason = "provider_disabled"
	RejectProviderNotApproved    RejectionReason = "provider_not_approved"
	RejectRouteDisabled          RejectionReason = "route_disabled"
	RejectRouteBlocked           RejectionReason = "route_blocked"
	RejectRouteDegraded          RejectionReason = "route_degraded"
	RejectUnsupportedCorridor    RejectionReason = "unsupported_corridor"
	RejectUnsupportedPayout      RejectionReason = "unsupported_payout_method"
	RejectUnsupportedCurrency    RejectionReason = "unsupported_currency"
	RejectUnsupportedBank        RejectionReason = "unsupported_destination_bank"
	RejectAmountOutsideLimits    RejectionReason = "amount_outside_limits"
	RejectStaleFXRate            RejectionReason = "stale_fx_rate"
	RejectInsufficientLiquidity  RejectionReason = "insufficient_liquidity"
	RejectComplianceManualReview RejectionReason = "compliance_manual_review"
	RejectCircuitBreakerOpen     RejectionReason = "circuit_breaker_open"
	RejectBelowMinSuccessRate    RejectionReason = "below_minimum_success_rate"
)

type RejectedRoute struct {
	RouteID    RouteID         `json:"route_id"`
	ProviderID ProviderID      `json:"provider_id"`
	Reason     RejectionReason `json:"reason"`
}

type CandidateRouteScore struct {
	RouteID         RouteID           `json:"route_id"`
	ProviderID      ProviderID        `json:"provider_id"`
	TotalScoreBps   uint16            `json:"total_score_bps"`
	ComponentScores map[string]uint16 `json:"component_scores"`
	PrimaryReason   string            `json:"primary_reason"`
}

type RouteDecision struct {
	TransactionID        string                `json:"transaction_id"`
	PolicyID             string                `json:"policy_id"`
	PolicyVersion        uint32                `json:"policy_version"`
	SelectedRoute        *CandidateRouteScore  `json:"selected_route,omitempty"`
	EligibleRoutes       []CandidateRouteScore `json:"eligible_routes"`
	RejectedRoutes       []RejectedRoute       `json:"rejected_routes"`
	DecidedAt            time.Time             `json:"decided_at"`
	AutoSwitchingAllowed bool                  `json:"auto_switching_allowed"`
}

func SelectRoute(registry RouteRegistry, policy BankPolicy, health RouteHealthProvider, input EligibilityInput) RouteDecision {
	eligibleRoutes := make([]CandidateRouteScore, 0)
	rejectedRoutes := make([]RejectedRoute, 0)

	for _, route := range registry.Routes {
		if route.Corridor != input.Corridor {
			rejectedRoutes = append(rejectedRoutes, reject(route, RejectUnsupportedCorridor))
			continue
		}

		if reason, rejected := rejectionReason(route, registry, policy, health, input); rejected {
			rejectedRoutes = append(rejectedRoutes, reject(route, reason))
			continue
		}

		eligibleRoutes = append(eligibleRoutes, scoreRoute(route, policy, health.SnapshotFor(route.ID)))
	}

	slices.SortFunc(eligibleRoutes, func(left, right CandidateRouteScore) int {
		if scoreOrder := cmp.Compare(right.TotalScoreBps, left.TotalScoreBps); scoreOrder != 0 {
			return scoreOrder
		}
		return cmp.Compare(left.RouteID, right.RouteID)
	})

	var selected *CandidateRouteScore
	if len(eligibleRoutes) > 0 {
		value := eligibleRoutes[0]
		selected = &value
	}

	return RouteDecision{
		TransactionID:        input.TransactionID,
		PolicyID:             policy.PolicyID,
		PolicyVersion:        policy.Version,
		SelectedRoute:        selected,
		EligibleRoutes:       eligibleRoutes,
		RejectedRoutes:       rejectedRoutes,
		DecidedAt:            time.Now().UTC(),
		AutoSwitchingAllowed: policy.AutoSwitchingAllowed,
	}
}

func rejectionReason(route Route, registry RouteRegistry, policy BankPolicy, health RouteHealthProvider, input EligibilityInput) (RejectionReason, bool) {
	provider, ok := registry.Provider(route.ProviderID)
	if !ok {
		return RejectProviderMissing, true
	}

	if !provider.Enabled || policy.DisabledProviders[provider.ID] {
		return RejectProviderDisabled, true
	}

	if !provider.Approved {
		return RejectProviderNotApproved, true
	}

	if policy.DisabledRoutes[route.ID] {
		return RejectRouteDisabled, true
	}

	if input.CircuitBreakerBlockedRoutes[route.ID] {
		return RejectCircuitBreakerOpen, true
	}

	switch route.Status {
	case RouteEnabled, RouteRecoveryTesting:
	case RouteDisabled:
		return RejectRouteDisabled, true
	case RouteDegraded:
		return RejectRouteDegraded, true
	case RouteBlocked:
		return RejectRouteBlocked, true
	}

	if route.PayoutMethod != input.PayoutMethod {
		return RejectUnsupportedPayout, true
	}

	if route.SettlementCurrency != input.Amount.Currency {
		return RejectUnsupportedCurrency, true
	}

	if !route.AmountRange.Contains(input.Amount) {
		return RejectAmountOutsideLimits, true
	}

	if !route.SupportsDestinationBank(input.DestinationBank) {
		return RejectUnsupportedBank, true
	}

	if input.ComplianceManualReviewRequired {
		return RejectComplianceManualReview, true
	}

	if !route.LiquidityAvailable {
		return RejectInsufficientLiquidity, true
	}

	if policy.MaxFXAgeSeconds != nil {
		if route.FXRateAgeSeconds == nil || *route.FXRateAgeSeconds > *policy.MaxFXAgeSeconds {
			return RejectStaleFXRate, true
		}
	}

	if policy.MinSuccessRateBps != nil {
		if health.SnapshotFor(route.ID).SuccessRateBps < *policy.MinSuccessRateBps {
			return RejectBelowMinSuccessRate, true
		}
	}

	return "", false
}

func reject(route Route, reason RejectionReason) RejectedRoute {
	return RejectedRoute{
		RouteID:    route.ID,
		ProviderID: route.ProviderID,
		Reason:     reason,
	}
}

func scoreRoute(route Route, policy BankPolicy, health RouteHealthSnapshot) CandidateRouteScore {
	reliability := averageBps(health.SuccessRateBps, health.UptimeBps, health.SettlementReliability)
	speed := latencyScore(health.P95LatencyMS, policy.TargetP95LatencyMS)
	cost := uint16(10_000 - min(route.CostPenaltyBps, 10_000))
	fx := min(route.FXQualityBps, 10_000)
	liquidity := uint16(0)
	if route.LiquidityAvailable {
		liquidity = 10_000
	}
	operations := uint16(10_000 - min(health.ManualInterventionBps, 10_000))

	weights := policy.Weights
	totalWeight := max(weights.Total(), 1)
	total := (uint32(reliability)*uint32(weights.Reliability) +
		uint32(speed)*uint32(weights.Speed) +
		uint32(cost)*uint32(weights.Cost) +
		uint32(fx)*uint32(weights.FX) +
		uint32(liquidity)*uint32(weights.Liquidity) +
		uint32(operations)*uint32(weights.Operations)) / totalWeight

	return CandidateRouteScore{
		RouteID:       route.ID,
		ProviderID:    route.ProviderID,
		TotalScoreBps: uint16(total),
		ComponentScores: map[string]uint16{
			"reliability": reliability,
			"speed":       speed,
			"cost":        cost,
			"fx":          fx,
			"liquidity":   liquidity,
			"operations":  operations,
		},
		PrimaryReason: primaryReason(reliability, speed, cost, fx),
	}
}

func averageBps(values ...uint16) uint16 {
	sum := uint32(0)
	for _, value := range values {
		sum += uint32(value)
	}
	return uint16(sum / uint32(len(values)))
}

func latencyScore(p95LatencyMS, targetP95LatencyMS uint64) uint16 {
	if p95LatencyMS <= targetP95LatencyMS {
		return 10_000
	}

	overshoot := p95LatencyMS - targetP95LatencyMS
	penalty := uint16(min((overshoot*10_000)/max(targetP95LatencyMS, 1), 10_000))
	return 10_000 - penalty
}

func primaryReason(reliability, speed, cost, fx uint16) string {
	parts := []struct {
		name  string
		score uint16
	}{
		{name: "reliability", score: reliability},
		{name: "speed", score: speed},
		{name: "cost", score: cost},
		{name: "fx", score: fx},
	}
	slices.SortFunc(parts, func(left, right struct {
		name  string
		score uint16
	}) int {
		if scoreOrder := cmp.Compare(right.score, left.score); scoreOrder != 0 {
			return scoreOrder
		}
		return cmp.Compare(left.name, right.name)
	})
	return "strongest " + parts[0].name
}
