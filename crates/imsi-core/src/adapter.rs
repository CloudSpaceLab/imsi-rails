use crate::lifecycle::{EventSource, TransactionEvent, TransactionState};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct SandboxSubmission {
    pub transaction_id: String,
    pub provider_reference: String,
    pub route_id: String,
}

#[derive(Clone, Copy, Debug, Deserialize, Eq, PartialEq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum SandboxOutcome {
    Accepted,
    Rejected,
    Timeout,
    Failed,
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct SandboxSubmissionResult {
    pub outcome: SandboxOutcome,
    pub simulated_latency_ms: u64,
    pub provider_reference: String,
    pub callbacks: Vec<SandboxCallback>,
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct SandboxCallback {
    pub transaction_id: String,
    pub provider_reference: String,
    pub state: TransactionState,
    pub occurred_at: DateTime<Utc>,
    pub duplicate: bool,
}

impl SandboxCallback {
    pub fn as_event(&self, event_id: impl Into<String>) -> TransactionEvent {
        TransactionEvent::new(
            event_id,
            self.transaction_id.clone(),
            self.state,
            EventSource::Provider,
            if self.duplicate {
                "duplicate sandbox provider callback"
            } else {
                "sandbox provider callback"
            },
        )
    }
}

#[derive(Clone, Debug, Deserialize, Eq, PartialEq, Serialize)]
pub struct SandboxProviderAdapter {
    pub provider_id: String,
    pub latency_ms: u64,
    pub failure_rate_bps: u16,
    pub rejection_rate_bps: u16,
    pub timeout_rate_bps: u16,
    pub duplicate_callbacks: u8,
}

impl SandboxProviderAdapter {
    pub fn always_accept(provider_id: impl Into<String>) -> Self {
        Self {
            provider_id: provider_id.into(),
            latency_ms: 0,
            failure_rate_bps: 0,
            rejection_rate_bps: 0,
            timeout_rate_bps: 0,
            duplicate_callbacks: 0,
        }
    }

    pub fn submit(&self, submission: SandboxSubmission) -> SandboxSubmissionResult {
        let bucket = deterministic_bucket(&submission.transaction_id);
        let timeout_cutoff = self.timeout_rate_bps.min(10_000);
        let rejection_cutoff =
            timeout_cutoff + self.rejection_rate_bps.min(10_000 - timeout_cutoff);
        let failure_cutoff =
            rejection_cutoff + self.failure_rate_bps.min(10_000 - rejection_cutoff);

        let outcome = if bucket < timeout_cutoff {
            SandboxOutcome::Timeout
        } else if bucket < rejection_cutoff {
            SandboxOutcome::Rejected
        } else if bucket < failure_cutoff {
            SandboxOutcome::Failed
        } else {
            SandboxOutcome::Accepted
        };

        SandboxSubmissionResult {
            outcome,
            simulated_latency_ms: self.latency_ms,
            provider_reference: submission.provider_reference.clone(),
            callbacks: callbacks_for(&submission, outcome, self.duplicate_callbacks),
        }
    }
}

fn callbacks_for(
    submission: &SandboxSubmission,
    outcome: SandboxOutcome,
    duplicate_callbacks: u8,
) -> Vec<SandboxCallback> {
    let state = match outcome {
        SandboxOutcome::Accepted => TransactionState::ProviderAccepted,
        SandboxOutcome::Rejected | SandboxOutcome::Failed | SandboxOutcome::Timeout => {
            TransactionState::Failed
        }
    };

    let callback = SandboxCallback {
        transaction_id: submission.transaction_id.clone(),
        provider_reference: submission.provider_reference.clone(),
        state,
        occurred_at: Utc::now(),
        duplicate: false,
    };

    let mut callbacks = vec![callback.clone()];
    for _ in 0..duplicate_callbacks {
        let mut duplicate = callback.clone();
        duplicate.duplicate = true;
        callbacks.push(duplicate);
    }

    callbacks
}

fn deterministic_bucket(value: &str) -> u16 {
    let hash = value.bytes().fold(0u32, |acc, byte| {
        acc.wrapping_mul(31).wrapping_add(u32::from(byte))
    });
    (hash % 10_000) as u16
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn sandbox_adapter_can_simulate_acceptance_and_duplicate_callbacks() {
        let adapter = SandboxProviderAdapter {
            provider_id: "sandbox".into(),
            latency_ms: 250,
            failure_rate_bps: 0,
            rejection_rate_bps: 0,
            timeout_rate_bps: 0,
            duplicate_callbacks: 2,
        };

        let result = adapter.submit(SandboxSubmission {
            transaction_id: "txn-1".into(),
            provider_reference: "sandbox-ref-1".into(),
            route_id: "route-1".into(),
        });

        assert_eq!(result.outcome, SandboxOutcome::Accepted);
        assert_eq!(result.simulated_latency_ms, 250);
        assert_eq!(result.callbacks.len(), 3);
        assert_eq!(
            result.callbacks[0].state,
            TransactionState::ProviderAccepted
        );
        assert!(result.callbacks[1].duplicate);
    }

    #[test]
    fn sandbox_adapter_can_force_timeout() {
        let adapter = SandboxProviderAdapter {
            provider_id: "sandbox".into(),
            latency_ms: 10_000,
            failure_rate_bps: 0,
            rejection_rate_bps: 0,
            timeout_rate_bps: 10_000,
            duplicate_callbacks: 0,
        };

        let result = adapter.submit(SandboxSubmission {
            transaction_id: "txn-timeout".into(),
            provider_reference: "sandbox-ref-2".into(),
            route_id: "route-1".into(),
        });

        assert_eq!(result.outcome, SandboxOutcome::Timeout);
        assert_eq!(result.callbacks[0].state, TransactionState::Failed);
    }
}
