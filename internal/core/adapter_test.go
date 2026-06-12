package core

import "testing"

func TestSandboxAdapterSimulatesAcceptanceAndDuplicateCallbacks(t *testing.T) {
	t.Parallel()

	adapter := SandboxProviderAdapter{
		ProviderID:         "sandbox",
		LatencyMS:          250,
		DuplicateCallbacks: 2,
	}

	result := adapter.Submit(SandboxSubmission{
		TransactionID:     "txn-1",
		ProviderReference: "sandbox-ref-1",
		RouteID:           "route-1",
	})

	if result.Outcome != SandboxAccepted {
		t.Fatalf("expected accepted, got %s", result.Outcome)
	}
	if result.SimulatedLatencyMS != 250 {
		t.Fatalf("expected latency 250, got %d", result.SimulatedLatencyMS)
	}
	if len(result.Callbacks) != 3 {
		t.Fatalf("expected 3 callbacks, got %d", len(result.Callbacks))
	}
	if result.Callbacks[0].State != StateProviderAccepted {
		t.Fatalf("expected provider_accepted, got %s", result.Callbacks[0].State)
	}
	if !result.Callbacks[1].Duplicate {
		t.Fatal("expected duplicate callback")
	}
}

func TestSandboxAdapterCanForceTimeout(t *testing.T) {
	t.Parallel()

	adapter := SandboxProviderAdapter{
		ProviderID:     "sandbox",
		LatencyMS:      10_000,
		TimeoutRateBps: 10_000,
	}

	result := adapter.Submit(SandboxSubmission{
		TransactionID:     "txn-timeout",
		ProviderReference: "sandbox-ref-2",
		RouteID:           "route-1",
	})

	if result.Outcome != SandboxTimeout {
		t.Fatalf("expected timeout, got %s", result.Outcome)
	}
	if result.Callbacks[0].State != StateFailed {
		t.Fatalf("expected failed callback, got %s", result.Callbacks[0].State)
	}
}
