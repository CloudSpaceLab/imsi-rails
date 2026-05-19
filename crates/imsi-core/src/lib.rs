//! Core domain model and routing primitives for imsi-rails.
//!
//! This crate intentionally keeps the routing hot path small: route capability,
//! policy, health, and FX freshness are evaluated from in-memory snapshots, while
//! callers persist the resulting decision for audit.

pub mod adapter;
pub mod audit;
pub mod lifecycle;
pub mod registry;
pub mod routing;

pub use adapter::{
    SandboxCallback, SandboxOutcome, SandboxProviderAdapter, SandboxSubmission,
    SandboxSubmissionResult,
};
pub use audit::{InMemoryRouteDecisionStore, RouteDecisionStore};
pub use lifecycle::{EventSource, TransactionEvent, TransactionState};
pub use registry::{
    AmountRange, Corridor, Money, PayoutMethod, Provider, ProviderId, Route, RouteId,
    RouteRegistry, RouteStatus,
};
pub use routing::{
    BankPolicy, CandidateRouteScore, EligibilityInput, RejectedRoute, RejectionReason,
    RouteDecision, RouteHealthBook, RouteHealthSnapshot, ScoringWeights, select_route,
};
