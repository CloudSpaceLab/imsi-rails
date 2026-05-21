package core

import "testing"

func TestRouteDecisionStoreRecordsAndRetrievesDecision(t *testing.T) {
	t.Parallel()

	store := NewInMemoryRouteDecisionStore()
	decision := RouteDecision{
		TransactionID:        "txn-1",
		PolicyID:             "default",
		PolicyVersion:        1,
		AutoSwitchingAllowed: true,
	}

	store.Record(decision)

	saved, ok := store.GetByTransactionID("txn-1")
	if !ok {
		t.Fatal("expected saved decision")
	}
	if saved.TransactionID != "txn-1" {
		t.Fatalf("expected txn-1, got %s", saved.TransactionID)
	}
	if _, ok := store.GetByTransactionID("missing"); ok {
		t.Fatal("did not expect missing decision")
	}
}
