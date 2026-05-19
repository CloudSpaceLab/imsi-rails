use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;

pub type ProviderId = String;
pub type RouteId = String;

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct Money {
    pub currency: String,
    pub minor_units: u64,
}

impl Money {
    pub fn new(currency: impl Into<String>, minor_units: u64) -> Self {
        Self {
            currency: currency.into().to_uppercase(),
            minor_units,
        }
    }
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct AmountRange {
    pub min_minor_units: u64,
    pub max_minor_units: u64,
}

impl AmountRange {
    pub fn contains(&self, amount: &Money) -> bool {
        amount.minor_units >= self.min_minor_units && amount.minor_units <= self.max_minor_units
    }
}

#[derive(Clone, Debug, Deserialize, Eq, Ord, PartialEq, PartialOrd, Serialize)]
pub struct Corridor {
    pub origin_country: String,
    pub destination_country: String,
}

impl Corridor {
    pub fn new(origin_country: impl Into<String>, destination_country: impl Into<String>) -> Self {
        Self {
            origin_country: origin_country.into().to_uppercase(),
            destination_country: destination_country.into().to_uppercase(),
        }
    }
}

#[derive(Clone, Copy, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum PayoutMethod {
    BankAccount,
    CashPickup,
    Wallet,
    Card,
    Branch,
}

#[derive(Clone, Copy, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum RouteStatus {
    Enabled,
    Disabled,
    Degraded,
    Blocked,
    RecoveryTesting,
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct Provider {
    pub id: ProviderId,
    pub name: String,
    pub enabled: bool,
    pub approved: bool,
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct Route {
    pub id: RouteId,
    pub provider_id: ProviderId,
    pub corridor: Corridor,
    pub payout_method: PayoutMethod,
    pub amount_range: AmountRange,
    pub settlement_currency: String,
    pub destination_banks: Vec<String>,
    pub status: RouteStatus,
    pub liquidity_available: bool,
    pub fx_rate_age_seconds: Option<u64>,
    /// Lower is cheaper. Basis points above the best known route for the same corridor.
    pub cost_penalty_bps: u16,
    /// Higher is better. Includes rate quality and spread competitiveness.
    pub fx_quality_bps: u16,
}

impl Route {
    pub fn supports_destination_bank(&self, destination_bank: Option<&str>) -> bool {
        if self.destination_banks.is_empty() {
            return true;
        }

        destination_bank.is_some_and(|bank| {
            let bank = bank.to_uppercase();
            self.destination_banks.iter().any(|item| item == &bank)
        })
    }
}

#[derive(Clone, Debug, Default, Deserialize, Eq, PartialEq, Serialize)]
pub struct RouteRegistry {
    pub providers: BTreeMap<ProviderId, Provider>,
    pub routes: BTreeMap<RouteId, Route>,
}

impl RouteRegistry {
    pub fn add_provider(&mut self, provider: Provider) {
        self.providers.insert(provider.id.clone(), provider);
    }

    pub fn add_route(&mut self, route: Route) {
        self.routes.insert(route.id.clone(), route);
    }

    pub fn provider(&self, provider_id: &str) -> Option<&Provider> {
        self.providers.get(provider_id)
    }

    pub fn routes_for_corridor<'a>(
        &'a self,
        corridor: &'a Corridor,
    ) -> impl Iterator<Item = &'a Route> + 'a {
        self.routes
            .values()
            .filter(move |route| &route.corridor == corridor)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn normalizes_money_and_corridor_codes() {
        let money = Money::new("usd", 10_000);
        let corridor = Corridor::new("us", "ng");

        assert_eq!(money.currency, "USD");
        assert_eq!(corridor.origin_country, "US");
        assert_eq!(corridor.destination_country, "NG");
    }

    #[test]
    fn matches_destination_bank_when_limited() {
        let route = Route {
            id: "route-1".into(),
            provider_id: "provider-1".into(),
            corridor: Corridor::new("US", "NG"),
            payout_method: PayoutMethod::BankAccount,
            amount_range: AmountRange {
                min_minor_units: 1,
                max_minor_units: 1_000_000,
            },
            settlement_currency: "NGN".into(),
            destination_banks: vec!["ZENITH".into()],
            status: RouteStatus::Enabled,
            liquidity_available: true,
            fx_rate_age_seconds: Some(10),
            cost_penalty_bps: 0,
            fx_quality_bps: 9_000,
        };

        assert!(route.supports_destination_bank(Some("zenith")));
        assert!(!route.supports_destination_bank(Some("gtbank")));
        assert!(!route.supports_destination_bank(None));
    }
}
