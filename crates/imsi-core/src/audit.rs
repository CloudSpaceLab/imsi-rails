use crate::routing::RouteDecision;
use std::collections::BTreeMap;
use std::sync::{Arc, RwLock};

pub trait RouteDecisionStore {
    fn record(&self, decision: RouteDecision);
    fn get_by_transaction_id(&self, transaction_id: &str) -> Option<RouteDecision>;
}

#[derive(Clone, Default)]
pub struct InMemoryRouteDecisionStore {
    decisions: Arc<RwLock<BTreeMap<String, RouteDecision>>>,
}

impl RouteDecisionStore for InMemoryRouteDecisionStore {
    fn record(&self, decision: RouteDecision) {
        let mut decisions = self
            .decisions
            .write()
            .expect("decision store lock poisoned");
        decisions.insert(decision.transaction_id.clone(), decision);
    }

    fn get_by_transaction_id(&self, transaction_id: &str) -> Option<RouteDecision> {
        let decisions = self.decisions.read().expect("decision store lock poisoned");
        decisions.get(transaction_id).cloned()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::routing::BankPolicy;
    use chrono::Utc;

    #[test]
    fn records_and_retrieves_decision_by_transaction_id() {
        let store = InMemoryRouteDecisionStore::default();
        let decision = RouteDecision {
            transaction_id: "txn-1".into(),
            policy_id: BankPolicy::default().policy_id,
            policy_version: 1,
            selected_route: None,
            eligible_routes: vec![],
            rejected_routes: vec![],
            decided_at: Utc::now(),
            auto_switching_allowed: true,
        };

        store.record(decision);
        let saved = store.get_by_transaction_id("txn-1");

        assert!(saved.is_some());
        assert_eq!(saved.unwrap().transaction_id, "txn-1");
        assert!(store.get_by_transaction_id("missing").is_none());
    }
}
