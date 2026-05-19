use crate::registry::{Corridor, Money, PayoutMethod, Route, RouteId, RouteRegistry, RouteStatus};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::{BTreeMap, BTreeSet};

#[derive(Clone, Copy, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct ScoringWeights {
    pub reliability: u16,
    pub speed: u16,
    pub cost: u16,
    pub fx: u16,
    pub liquidity: u16,
    pub operations: u16,
}

impl Default for ScoringWeights {
    fn default() -> Self {
        Self {
            reliability: 35,
            speed: 20,
            cost: 15,
            fx: 10,
            liquidity: 10,
            operations: 10,
        }
    }
}

impl ScoringWeights {
    pub fn total(self) -> u32 {
        u32::from(self.reliability)
            + u32::from(self.speed)
            + u32::from(self.cost)
            + u32::from(self.fx)
            + u32::from(self.liquidity)
            + u32::from(self.operations)
    }
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct BankPolicy {
    pub policy_id: String,
    pub version: u32,
    pub weights: ScoringWeights,
    pub disabled_providers: BTreeSet<String>,
    pub disabled_routes: BTreeSet<RouteId>,
    pub max_fx_age_seconds: Option<u64>,
    pub min_success_rate_bps: Option<u16>,
    pub target_p95_latency_ms: u64,
    pub auto_switching_allowed: bool,
}

impl Default for BankPolicy {
    fn default() -> Self {
        Self {
            policy_id: "default".into(),
            version: 1,
            weights: ScoringWeights::default(),
            disabled_providers: BTreeSet::new(),
            disabled_routes: BTreeSet::new(),
            max_fx_age_seconds: Some(180),
            min_success_rate_bps: Some(8_000),
            target_p95_latency_ms: 120_000,
            auto_switching_allowed: true,
        }
    }
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct EligibilityInput {
    pub transaction_id: String,
    pub corridor: Corridor,
    pub payout_method: PayoutMethod,
    pub amount: Money,
    pub destination_bank: Option<String>,
    pub compliance_manual_review_required: bool,
    pub circuit_breaker_blocked_routes: BTreeSet<RouteId>,
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct RouteHealthSnapshot {
    pub success_rate_bps: u16,
    pub uptime_bps: u16,
    pub p95_latency_ms: u64,
    pub manual_intervention_bps: u16,
    pub settlement_reliability_bps: u16,
    pub observed_at: DateTime<Utc>,
}

impl Default for RouteHealthSnapshot {
    fn default() -> Self {
        Self {
            success_rate_bps: 10_000,
            uptime_bps: 10_000,
            p95_latency_ms: 1_000,
            manual_intervention_bps: 0,
            settlement_reliability_bps: 10_000,
            observed_at: Utc::now(),
        }
    }
}

#[derive(Clone, Debug, Default, Deserialize, Eq, PartialEq, Serialize)]
pub struct RouteHealthBook {
    pub routes: BTreeMap<RouteId, RouteHealthSnapshot>,
}

impl RouteHealthBook {
    pub fn snapshot_for(&self, route_id: &str) -> RouteHealthSnapshot {
        self.routes.get(route_id).cloned().unwrap_or_default()
    }
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum RejectionReason {
    ProviderMissing,
    ProviderDisabled,
    ProviderNotApproved,
    RouteDisabled,
    RouteBlocked,
    RouteDegraded,
    UnsupportedCorridor,
    UnsupportedPayoutMethod,
    UnsupportedCurrency,
    UnsupportedDestinationBank,
    AmountOutsideLimits,
    StaleFxRate,
    InsufficientLiquidity,
    ComplianceManualReview,
    CircuitBreakerOpen,
    BelowMinimumSuccessRate,
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct RejectedRoute {
    pub route_id: RouteId,
    pub provider_id: String,
    pub reason: RejectionReason,
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct CandidateRouteScore {
    pub route_id: RouteId,
    pub provider_id: String,
    /// 0-10000 where 10000 is best.
    pub total_score_bps: u16,
    pub component_scores: BTreeMap<String, u16>,
    pub primary_reason: String,
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct RouteDecision {
    pub transaction_id: String,
    pub policy_id: String,
    pub policy_version: u32,
    pub selected_route: Option<CandidateRouteScore>,
    pub eligible_routes: Vec<CandidateRouteScore>,
    pub rejected_routes: Vec<RejectedRoute>,
    pub decided_at: DateTime<Utc>,
    pub auto_switching_allowed: bool,
}

pub fn select_route(
    registry: &RouteRegistry,
    policy: &BankPolicy,
    health: &RouteHealthBook,
    input: EligibilityInput,
) -> RouteDecision {
    let mut eligible_routes = Vec::new();
    let mut rejected_routes = Vec::new();

    for route in registry.routes.values() {
        if route.corridor != input.corridor {
            rejected_routes.push(reject(route, RejectionReason::UnsupportedCorridor));
            continue;
        }

        if let Some(reason) = rejection_reason(route, registry, policy, health, &input) {
            rejected_routes.push(reject(route, reason));
            continue;
        }

        eligible_routes.push(score_route(route, policy, &health.snapshot_for(&route.id)));
    }

    eligible_routes.sort_by(|left, right| {
        right
            .total_score_bps
            .cmp(&left.total_score_bps)
            .then_with(|| left.route_id.cmp(&right.route_id))
    });

    RouteDecision {
        transaction_id: input.transaction_id,
        policy_id: policy.policy_id.clone(),
        policy_version: policy.version,
        selected_route: eligible_routes.first().cloned(),
        eligible_routes,
        rejected_routes,
        decided_at: Utc::now(),
        auto_switching_allowed: policy.auto_switching_allowed,
    }
}

fn rejection_reason(
    route: &Route,
    registry: &RouteRegistry,
    policy: &BankPolicy,
    health: &RouteHealthBook,
    input: &EligibilityInput,
) -> Option<RejectionReason> {
    let provider = match registry.provider(&route.provider_id) {
        Some(provider) => provider,
        None => return Some(RejectionReason::ProviderMissing),
    };

    if !provider.enabled || policy.disabled_providers.contains(&provider.id) {
        return Some(RejectionReason::ProviderDisabled);
    }

    if !provider.approved {
        return Some(RejectionReason::ProviderNotApproved);
    }

    if policy.disabled_routes.contains(&route.id) {
        return Some(RejectionReason::RouteDisabled);
    }

    if input.circuit_breaker_blocked_routes.contains(&route.id) {
        return Some(RejectionReason::CircuitBreakerOpen);
    }

    match route.status {
        RouteStatus::Enabled | RouteStatus::RecoveryTesting => {}
        RouteStatus::Disabled => return Some(RejectionReason::RouteDisabled),
        RouteStatus::Degraded => return Some(RejectionReason::RouteDegraded),
        RouteStatus::Blocked => return Some(RejectionReason::RouteBlocked),
    }

    if route.payout_method != input.payout_method {
        return Some(RejectionReason::UnsupportedPayoutMethod);
    }

    if route.settlement_currency != input.amount.currency {
        return Some(RejectionReason::UnsupportedCurrency);
    }

    if !route.amount_range.contains(&input.amount) {
        return Some(RejectionReason::AmountOutsideLimits);
    }

    if !route.supports_destination_bank(input.destination_bank.as_deref()) {
        return Some(RejectionReason::UnsupportedDestinationBank);
    }

    if input.compliance_manual_review_required {
        return Some(RejectionReason::ComplianceManualReview);
    }

    if !route.liquidity_available {
        return Some(RejectionReason::InsufficientLiquidity);
    }

    if let Some(max_age) = policy.max_fx_age_seconds
        && route.fx_rate_age_seconds.is_none_or(|age| age > max_age)
    {
        return Some(RejectionReason::StaleFxRate);
    }

    if let Some(min_success) = policy.min_success_rate_bps {
        let snapshot = health.snapshot_for(&route.id);
        if snapshot.success_rate_bps < min_success {
            return Some(RejectionReason::BelowMinimumSuccessRate);
        }
    }

    None
}

fn reject(route: &Route, reason: RejectionReason) -> RejectedRoute {
    RejectedRoute {
        route_id: route.id.clone(),
        provider_id: route.provider_id.clone(),
        reason,
    }
}

fn score_route(
    route: &Route,
    policy: &BankPolicy,
    health: &RouteHealthSnapshot,
) -> CandidateRouteScore {
    let reliability = average_bps(&[
        health.success_rate_bps,
        health.uptime_bps,
        health.settlement_reliability_bps,
    ]);
    let speed = latency_score(health.p95_latency_ms, policy.target_p95_latency_ms);
    let cost = 10_000u16.saturating_sub(route.cost_penalty_bps.min(10_000));
    let fx = route.fx_quality_bps.min(10_000);
    let liquidity = if route.liquidity_available { 10_000 } else { 0 };
    let operations = 10_000u16.saturating_sub(health.manual_intervention_bps.min(10_000));

    let weights = policy.weights;
    let total_weight = weights.total().max(1);
    let total = (u32::from(reliability) * u32::from(weights.reliability)
        + u32::from(speed) * u32::from(weights.speed)
        + u32::from(cost) * u32::from(weights.cost)
        + u32::from(fx) * u32::from(weights.fx)
        + u32::from(liquidity) * u32::from(weights.liquidity)
        + u32::from(operations) * u32::from(weights.operations))
        / total_weight;

    let mut component_scores = BTreeMap::new();
    component_scores.insert("reliability".into(), reliability);
    component_scores.insert("speed".into(), speed);
    component_scores.insert("cost".into(), cost);
    component_scores.insert("fx".into(), fx);
    component_scores.insert("liquidity".into(), liquidity);
    component_scores.insert("operations".into(), operations);

    CandidateRouteScore {
        route_id: route.id.clone(),
        provider_id: route.provider_id.clone(),
        total_score_bps: total as u16,
        component_scores,
        primary_reason: primary_reason(reliability, speed, cost, fx),
    }
}

fn average_bps(values: &[u16]) -> u16 {
    let sum: u32 = values.iter().map(|value| u32::from(*value)).sum();
    (sum / values.len() as u32) as u16
}

fn latency_score(p95_latency_ms: u64, target_p95_latency_ms: u64) -> u16 {
    if p95_latency_ms <= target_p95_latency_ms {
        return 10_000;
    }

    let overshoot = p95_latency_ms - target_p95_latency_ms;
    let penalty = ((overshoot.saturating_mul(10_000)) / target_p95_latency_ms.max(1)) as u16;
    10_000u16.saturating_sub(penalty.min(10_000))
}

fn primary_reason(reliability: u16, speed: u16, cost: u16, fx: u16) -> String {
    let mut parts = [
        ("reliability", reliability),
        ("speed", speed),
        ("cost", cost),
        ("fx", fx),
    ];
    parts.sort_by(|left, right| right.1.cmp(&left.1).then_with(|| left.0.cmp(right.0)));
    format!("strongest {}", parts[0].0)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::registry::{AmountRange, Provider};

    fn route(id: &str, provider_id: &str, cost_penalty_bps: u16) -> Route {
        Route {
            id: id.into(),
            provider_id: provider_id.into(),
            corridor: Corridor::new("US", "NG"),
            payout_method: PayoutMethod::BankAccount,
            amount_range: AmountRange {
                min_minor_units: 1,
                max_minor_units: 50_000_000,
            },
            settlement_currency: "NGN".into(),
            destination_banks: vec![],
            status: RouteStatus::Enabled,
            liquidity_available: true,
            fx_rate_age_seconds: Some(30),
            cost_penalty_bps,
            fx_quality_bps: 9_000,
        }
    }

    fn input() -> EligibilityInput {
        EligibilityInput {
            transaction_id: "txn-1".into(),
            corridor: Corridor::new("US", "NG"),
            payout_method: PayoutMethod::BankAccount,
            amount: Money::new("NGN", 10_000),
            destination_bank: Some("ACCESS".into()),
            compliance_manual_review_required: false,
            circuit_breaker_blocked_routes: BTreeSet::new(),
        }
    }

    #[test]
    fn selects_highest_scored_eligible_route() {
        let mut registry = RouteRegistry::default();
        registry.add_provider(Provider {
            id: "p1".into(),
            name: "Provider 1".into(),
            enabled: true,
            approved: true,
        });
        registry.add_provider(Provider {
            id: "p2".into(),
            name: "Provider 2".into(),
            enabled: true,
            approved: true,
        });
        registry.add_route(route("r1", "p1", 2_000));
        registry.add_route(route("r2", "p2", 0));

        let mut health = RouteHealthBook::default();
        health.routes.insert(
            "r1".into(),
            RouteHealthSnapshot {
                success_rate_bps: 8_200,
                p95_latency_ms: 80_000,
                ..RouteHealthSnapshot::default()
            },
        );
        health.routes.insert(
            "r2".into(),
            RouteHealthSnapshot {
                success_rate_bps: 9_800,
                p95_latency_ms: 30_000,
                ..RouteHealthSnapshot::default()
            },
        );

        let decision = select_route(&registry, &BankPolicy::default(), &health, input());

        assert_eq!(decision.selected_route.unwrap().route_id, "r2");
        assert_eq!(decision.eligible_routes.len(), 2);
    }

    #[test]
    fn rejects_stale_fx_and_open_circuit_breaker() {
        let mut registry = RouteRegistry::default();
        registry.add_provider(Provider {
            id: "p1".into(),
            name: "Provider 1".into(),
            enabled: true,
            approved: true,
        });
        let mut stale = route("stale", "p1", 0);
        stale.fx_rate_age_seconds = Some(999);
        registry.add_route(stale);
        registry.add_route(route("blocked", "p1", 0));

        let mut input = input();
        input
            .circuit_breaker_blocked_routes
            .insert("blocked".into());

        let decision = select_route(
            &registry,
            &BankPolicy::default(),
            &RouteHealthBook::default(),
            input,
        );

        assert!(decision.selected_route.is_none());
        assert_eq!(decision.rejected_routes.len(), 2);
        assert!(
            decision
                .rejected_routes
                .iter()
                .any(|route| route.reason == RejectionReason::StaleFxRate)
        );
        assert!(
            decision
                .rejected_routes
                .iter()
                .any(|route| route.reason == RejectionReason::CircuitBreakerOpen)
        );
    }
}
