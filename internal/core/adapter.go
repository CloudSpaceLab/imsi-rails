package core

import "time"

type SandboxSubmission struct {
	TransactionID     string  `json:"transaction_id"`
	ProviderReference string  `json:"provider_reference"`
	RouteID           RouteID `json:"route_id"`
}

type SandboxOutcome string

const (
	SandboxAccepted SandboxOutcome = "accepted"
	SandboxRejected SandboxOutcome = "rejected"
	SandboxTimeout  SandboxOutcome = "timeout"
	SandboxFailed   SandboxOutcome = "failed"
)

type SandboxSubmissionResult struct {
	Outcome            SandboxOutcome    `json:"outcome"`
	SimulatedLatencyMS uint64            `json:"simulated_latency_ms"`
	ProviderReference  string            `json:"provider_reference"`
	Callbacks          []SandboxCallback `json:"callbacks"`
}

type SandboxCallback struct {
	TransactionID     string           `json:"transaction_id"`
	ProviderReference string           `json:"provider_reference"`
	State             TransactionState `json:"state"`
	OccurredAt        time.Time        `json:"occurred_at"`
	Duplicate         bool             `json:"duplicate"`
}

func (c SandboxCallback) AsEvent(eventID string) TransactionEvent {
	message := "sandbox provider callback"
	if c.Duplicate {
		message = "duplicate sandbox provider callback"
	}

	return NewTransactionEvent(eventID, c.TransactionID, c.State, EventSourceProvider, message)
}

type SandboxProviderAdapter struct {
	ProviderID         ProviderID `json:"provider_id"`
	LatencyMS          uint64     `json:"latency_ms"`
	FailureRateBps     uint16     `json:"failure_rate_bps"`
	RejectionRateBps   uint16     `json:"rejection_rate_bps"`
	TimeoutRateBps     uint16     `json:"timeout_rate_bps"`
	DuplicateCallbacks uint8      `json:"duplicate_callbacks"`
}

func NewAlwaysAcceptSandboxAdapter(providerID ProviderID) SandboxProviderAdapter {
	return SandboxProviderAdapter{ProviderID: providerID}
}

func (a SandboxProviderAdapter) Submit(submission SandboxSubmission) SandboxSubmissionResult {
	bucket := deterministicBucket(submission.TransactionID)
	timeoutCutoff := min(a.TimeoutRateBps, 10_000)
	rejectionCutoff := timeoutCutoff + min(a.RejectionRateBps, 10_000-timeoutCutoff)
	failureCutoff := rejectionCutoff + min(a.FailureRateBps, 10_000-rejectionCutoff)

	outcome := SandboxAccepted
	switch {
	case bucket < timeoutCutoff:
		outcome = SandboxTimeout
	case bucket < rejectionCutoff:
		outcome = SandboxRejected
	case bucket < failureCutoff:
		outcome = SandboxFailed
	}

	return SandboxSubmissionResult{
		Outcome:            outcome,
		SimulatedLatencyMS: a.LatencyMS,
		ProviderReference:  submission.ProviderReference,
		Callbacks:          callbacksFor(submission, outcome, a.DuplicateCallbacks),
	}
}

func callbacksFor(submission SandboxSubmission, outcome SandboxOutcome, duplicateCallbacks uint8) []SandboxCallback {
	state := StateProviderAccepted
	if outcome != SandboxAccepted {
		state = StateFailed
	}

	callback := SandboxCallback{
		TransactionID:     submission.TransactionID,
		ProviderReference: submission.ProviderReference,
		State:             state,
		OccurredAt:        time.Now().UTC(),
	}

	callbacks := []SandboxCallback{callback}
	for range duplicateCallbacks {
		duplicate := callback
		duplicate.Duplicate = true
		callbacks = append(callbacks, duplicate)
	}

	return callbacks
}

func deterministicBucket(value string) uint16 {
	hash := uint32(0)
	for _, b := range []byte(value) {
		hash = hash*31 + uint32(b)
	}
	return uint16(hash % 10_000)
}
