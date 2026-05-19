use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;

/// Canonical transaction states used across providers, banks, UI, and audit.
#[derive(Clone, Copy, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum TransactionState {
    Created,
    Accepted,
    CompliancePending,
    PrefundPending,
    SentToProvider,
    ProviderAccepted,
    ReceivedByBank,
    AccountValidated,
    PayoutSubmitted,
    Credited,
    PaidCash,
    Failed,
    Reversed,
    Disputed,
    Reconciled,
}

impl TransactionState {
    pub fn is_terminal(self) -> bool {
        matches!(
            self,
            Self::Failed | Self::Reversed | Self::Disputed | Self::Reconciled
        )
    }

    pub fn value_delivered(self) -> bool {
        matches!(self, Self::Credited | Self::PaidCash | Self::Reconciled)
    }

    /// Default safe boundary for automatic failover.
    ///
    /// Provider-specific adapters can be stricter, but the core should not
    /// automatically re-route once the transfer may have been accepted for
    /// payout by a provider or downstream bank.
    pub fn allows_auto_failover_by_default(self) -> bool {
        matches!(
            self,
            Self::Created | Self::Accepted | Self::CompliancePending | Self::PrefundPending
        )
    }

    pub fn can_transition_to(self, next: Self) -> bool {
        use TransactionState::*;

        matches!(
            (self, next),
            (Created, Accepted)
                | (Created, Failed)
                | (Accepted, CompliancePending)
                | (Accepted, PrefundPending)
                | (Accepted, SentToProvider)
                | (Accepted, Failed)
                | (CompliancePending, PrefundPending)
                | (CompliancePending, SentToProvider)
                | (CompliancePending, Failed)
                | (PrefundPending, SentToProvider)
                | (PrefundPending, Failed)
                | (SentToProvider, ProviderAccepted)
                | (SentToProvider, Failed)
                | (ProviderAccepted, ReceivedByBank)
                | (ProviderAccepted, Failed)
                | (ReceivedByBank, AccountValidated)
                | (ReceivedByBank, Failed)
                | (AccountValidated, PayoutSubmitted)
                | (AccountValidated, Failed)
                | (PayoutSubmitted, Credited)
                | (PayoutSubmitted, PaidCash)
                | (PayoutSubmitted, Failed)
                | (Credited, Reconciled)
                | (Credited, Disputed)
                | (PaidCash, Reconciled)
                | (PaidCash, Disputed)
                | (Failed, Reversed)
                | (Disputed, Reversed)
        )
    }
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum EventSource {
    Bank,
    Provider,
    Switch,
    Compliance,
    Settlement,
    Operator,
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct TransactionEvent {
    pub event_id: String,
    pub transaction_id: String,
    pub state: TransactionState,
    pub source: EventSource,
    pub occurred_at: DateTime<Utc>,
    pub message: String,
    pub references: BTreeMap<String, String>,
}

impl TransactionEvent {
    pub fn new(
        event_id: impl Into<String>,
        transaction_id: impl Into<String>,
        state: TransactionState,
        source: EventSource,
        message: impl Into<String>,
    ) -> Self {
        Self {
            event_id: event_id.into(),
            transaction_id: transaction_id.into(),
            state,
            source,
            occurred_at: Utc::now(),
            message: message.into(),
            references: BTreeMap::new(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::TransactionState::*;

    #[test]
    fn supports_expected_lifecycle_transitions() {
        assert!(Created.can_transition_to(Accepted));
        assert!(Accepted.can_transition_to(SentToProvider));
        assert!(SentToProvider.can_transition_to(ProviderAccepted));
        assert!(PayoutSubmitted.can_transition_to(Credited));
        assert!(Credited.can_transition_to(Reconciled));
        assert!(!Created.can_transition_to(Credited));
        assert!(!Reconciled.can_transition_to(Failed));
    }

    #[test]
    fn identifies_unsafe_auto_failover_boundary() {
        assert!(Accepted.allows_auto_failover_by_default());
        assert!(PrefundPending.allows_auto_failover_by_default());
        assert!(!SentToProvider.allows_auto_failover_by_default());
        assert!(!ProviderAccepted.allows_auto_failover_by_default());
        assert!(!Credited.allows_auto_failover_by_default());
    }
}
