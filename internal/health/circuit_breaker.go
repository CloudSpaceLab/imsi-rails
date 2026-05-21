package health

import (
	"time"

	"github.com/CloudSpaceLab/imsi-rails/internal/core"
)

type CircuitBreakerState string

const (
	BreakerUnknown         CircuitBreakerState = "unknown"
	BreakerHealthy         CircuitBreakerState = "healthy"
	BreakerDegraded        CircuitBreakerState = "degraded"
	BreakerBlocked         CircuitBreakerState = "blocked"
	BreakerRecoveryTesting CircuitBreakerState = "recovery_testing"
)

type CircuitBreakerThresholds struct {
	DegradedSuccessRateBps uint16 `json:"degraded_success_rate_bps"`
	BlockedSuccessRateBps  uint16 `json:"blocked_success_rate_bps"`
	DegradedUptimeBps      uint16 `json:"degraded_uptime_bps"`
	BlockedUptimeBps       uint16 `json:"blocked_uptime_bps"`
	DegradedP95LatencyMS   uint64 `json:"degraded_p95_latency_ms"`
	BlockedP95LatencyMS    uint64 `json:"blocked_p95_latency_ms"`
	RecoverySuccessRateBps uint16 `json:"recovery_success_rate_bps"`
	RecoveryUptimeBps      uint16 `json:"recovery_uptime_bps"`
	RecoveryP95LatencyMS   uint64 `json:"recovery_p95_latency_ms"`
}

func DefaultCircuitBreakerThresholds() CircuitBreakerThresholds {
	return CircuitBreakerThresholds{
		DegradedSuccessRateBps: 9_500,
		BlockedSuccessRateBps:  8_000,
		DegradedUptimeBps:      9_800,
		BlockedUptimeBps:       9_000,
		DegradedP95LatencyMS:   60_000,
		BlockedP95LatencyMS:    120_000,
		RecoverySuccessRateBps: 9_700,
		RecoveryUptimeBps:      9_900,
		RecoveryP95LatencyMS:   45_000,
	}
}

type CircuitBreakerRecord struct {
	RouteID    core.RouteID             `json:"route_id"`
	State      CircuitBreakerState      `json:"state"`
	Reason     string                   `json:"reason"`
	Snapshot   core.RouteHealthSnapshot `json:"snapshot"`
	Thresholds CircuitBreakerThresholds `json:"thresholds"`
	UpdatedAt  time.Time                `json:"updated_at"`
}

type CircuitBreakerStateChangeEvent struct {
	EventID       string              `json:"event_id"`
	RouteID       core.RouteID        `json:"route_id"`
	PreviousState CircuitBreakerState `json:"previous_state"`
	CurrentState  CircuitBreakerState `json:"current_state"`
	OccurredAt    time.Time           `json:"occurred_at"`
	Reason        string              `json:"reason"`
}

func EvaluateCircuitBreaker(routeID core.RouteID, previous CircuitBreakerRecord, snapshot core.RouteHealthSnapshot, thresholds CircuitBreakerThresholds, observedAt time.Time) CircuitBreakerRecord {
	rawState, reason := stateFromSnapshot(snapshot, thresholds)
	nextState := rawState

	switch previous.State {
	case BreakerBlocked:
		if meetsRecoveryThreshold(snapshot, thresholds) {
			nextState = BreakerRecoveryTesting
			reason = "recovery_threshold_met"
		} else {
			nextState = BreakerBlocked
			if reason == "within_threshold" {
				reason = "waiting_for_recovery_threshold"
			}
		}
	case BreakerRecoveryTesting:
		switch rawState {
		case BreakerHealthy:
			nextState = BreakerHealthy
			reason = "recovery_passed"
		case BreakerBlocked:
			nextState = BreakerBlocked
		default:
			nextState = BreakerRecoveryTesting
			reason = "recovery_observing"
		}
	}

	return CircuitBreakerRecord{
		RouteID:    routeID,
		State:      nextState,
		Reason:     reason,
		Snapshot:   snapshot,
		Thresholds: thresholds,
		UpdatedAt:  observedAt,
	}
}

func stateFromSnapshot(snapshot core.RouteHealthSnapshot, thresholds CircuitBreakerThresholds) (CircuitBreakerState, string) {
	switch {
	case snapshot.SuccessRateBps <= thresholds.BlockedSuccessRateBps:
		return BreakerBlocked, "success_rate_below_blocked_threshold"
	case snapshot.UptimeBps <= thresholds.BlockedUptimeBps:
		return BreakerBlocked, "uptime_below_blocked_threshold"
	case snapshot.P95LatencyMS >= thresholds.BlockedP95LatencyMS:
		return BreakerBlocked, "p95_latency_above_blocked_threshold"
	case snapshot.SuccessRateBps <= thresholds.DegradedSuccessRateBps:
		return BreakerDegraded, "success_rate_below_degraded_threshold"
	case snapshot.UptimeBps <= thresholds.DegradedUptimeBps:
		return BreakerDegraded, "uptime_below_degraded_threshold"
	case snapshot.P95LatencyMS >= thresholds.DegradedP95LatencyMS:
		return BreakerDegraded, "p95_latency_above_degraded_threshold"
	default:
		return BreakerHealthy, "within_threshold"
	}
}

func meetsRecoveryThreshold(snapshot core.RouteHealthSnapshot, thresholds CircuitBreakerThresholds) bool {
	return snapshot.SuccessRateBps >= thresholds.RecoverySuccessRateBps &&
		snapshot.UptimeBps >= thresholds.RecoveryUptimeBps &&
		snapshot.P95LatencyMS <= thresholds.RecoveryP95LatencyMS
}
