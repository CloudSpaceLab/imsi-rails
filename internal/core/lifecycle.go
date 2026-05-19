package core

import "time"

// TransactionState is the canonical state model shared by providers, banks,
// UI, analytics, and reconciliation.
type TransactionState string

const (
	StateCreated           TransactionState = "created"
	StateAccepted          TransactionState = "accepted"
	StateCompliancePending TransactionState = "compliance_pending"
	StatePrefundPending    TransactionState = "prefund_pending"
	StateSentToProvider    TransactionState = "sent_to_provider"
	StateProviderAccepted  TransactionState = "provider_accepted"
	StateReceivedByBank    TransactionState = "received_by_bank"
	StateAccountValidated  TransactionState = "account_validated"
	StatePayoutSubmitted   TransactionState = "payout_submitted"
	StateCredited          TransactionState = "credited"
	StatePaidCash          TransactionState = "paid_cash"
	StateFailed            TransactionState = "failed"
	StateReversed          TransactionState = "reversed"
	StateDisputed          TransactionState = "disputed"
	StateReconciled        TransactionState = "reconciled"
)

func (s TransactionState) IsTerminal() bool {
	switch s {
	case StateFailed, StateReversed, StateDisputed, StateReconciled:
		return true
	default:
		return false
	}
}

func (s TransactionState) ValueDelivered() bool {
	switch s {
	case StateCredited, StatePaidCash, StateReconciled:
		return true
	default:
		return false
	}
}

// AllowsAutoFailoverByDefault is deliberately conservative. Provider adapters
// can be stricter, but the core should not automatically re-route once the
// transaction may have been accepted for payout by a provider or downstream bank.
func (s TransactionState) AllowsAutoFailoverByDefault() bool {
	switch s {
	case StateCreated, StateAccepted, StateCompliancePending, StatePrefundPending:
		return true
	default:
		return false
	}
}

func (s TransactionState) CanTransitionTo(next TransactionState) bool {
	allowed := map[TransactionState][]TransactionState{
		StateCreated:           {StateAccepted, StateFailed},
		StateAccepted:          {StateCompliancePending, StatePrefundPending, StateSentToProvider, StateFailed},
		StateCompliancePending: {StatePrefundPending, StateSentToProvider, StateFailed},
		StatePrefundPending:    {StateSentToProvider, StateFailed},
		StateSentToProvider:    {StateProviderAccepted, StateFailed},
		StateProviderAccepted:  {StateReceivedByBank, StateFailed},
		StateReceivedByBank:    {StateAccountValidated, StateFailed},
		StateAccountValidated:  {StatePayoutSubmitted, StateFailed},
		StatePayoutSubmitted:   {StateCredited, StatePaidCash, StateFailed},
		StateCredited:          {StateReconciled, StateDisputed},
		StatePaidCash:          {StateReconciled, StateDisputed},
		StateFailed:            {StateReversed},
		StateDisputed:          {StateReversed},
	}

	for _, candidate := range allowed[s] {
		if candidate == next {
			return true
		}
	}
	return false
}

type EventSource string

const (
	EventSourceBank       EventSource = "bank"
	EventSourceProvider   EventSource = "provider"
	EventSourceSwitch     EventSource = "switch"
	EventSourceCompliance EventSource = "compliance"
	EventSourceSettlement EventSource = "settlement"
	EventSourceOperator   EventSource = "operator"
)

type TransactionEvent struct {
	EventID       string            `json:"event_id"`
	TransactionID string            `json:"transaction_id"`
	State         TransactionState  `json:"state"`
	Source        EventSource       `json:"source"`
	OccurredAt    time.Time         `json:"occurred_at"`
	Message       string            `json:"message"`
	References    map[string]string `json:"references"`
}

func NewTransactionEvent(eventID, transactionID string, state TransactionState, source EventSource, message string) TransactionEvent {
	return TransactionEvent{
		EventID:       eventID,
		TransactionID: transactionID,
		State:         state,
		Source:        source,
		OccurredAt:    time.Now().UTC(),
		Message:       message,
		References:    map[string]string{},
	}
}
