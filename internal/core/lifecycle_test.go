package core

import "testing"

func TestLifecycleTransitions(t *testing.T) {
	t.Parallel()

	if !StateCreated.CanTransitionTo(StateAccepted) {
		t.Fatal("created should transition to accepted")
	}
	if !StateSentToProvider.CanTransitionTo(StateProviderAccepted) {
		t.Fatal("sent_to_provider should transition to provider_accepted")
	}
	if !StatePayoutSubmitted.CanTransitionTo(StateCredited) {
		t.Fatal("payout_submitted should transition to credited")
	}
	if StateCreated.CanTransitionTo(StateCredited) {
		t.Fatal("created should not transition directly to credited")
	}
	if StateReconciled.CanTransitionTo(StateFailed) {
		t.Fatal("reconciled should not transition to failed")
	}
}

func TestAutoFailoverBoundary(t *testing.T) {
	t.Parallel()

	safeStates := []TransactionState{
		StateCreated,
		StateAccepted,
		StateCompliancePending,
		StatePrefundPending,
	}
	for _, state := range safeStates {
		if !state.AllowsAutoFailoverByDefault() {
			t.Fatalf("%s should allow default auto failover", state)
		}
	}

	unsafeStates := []TransactionState{
		StateSentToProvider,
		StateProviderAccepted,
		StateReceivedByBank,
		StateCredited,
	}
	for _, state := range unsafeStates {
		if state.AllowsAutoFailoverByDefault() {
			t.Fatalf("%s should not allow default auto failover", state)
		}
	}
}
