package intake

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/CloudSpaceLab/imsi-rails/internal/core"
)

var (
	ErrMissingIDempotencyKey = errors.New("idempotency_key is required")
	ErrMissingBankID         = errors.New("bank_id is required")
	ErrMissingCurrency       = errors.New("currency is required")
	ErrInvalidAmount         = errors.New("amount_minor_units must be greater than zero")
	ErrMissingPayoutMethod   = errors.New("payout_method is required")
	ErrMissingCountries      = errors.New("sender_country and receiver_country are required")
)

type SubmitTransactionRequest struct {
	BankID                         string            `json:"bank_id"`
	IDempotencyKey                 string            `json:"idempotency_key"`
	SenderCountry                  string            `json:"sender_country"`
	ReceiverCountry                string            `json:"receiver_country"`
	AmountMinorUnits               uint64            `json:"amount_minor_units"`
	Currency                       string            `json:"currency"`
	PayoutMethod                   core.PayoutMethod `json:"payout_method"`
	DestinationBank                string            `json:"destination_bank,omitempty"`
	ComplianceManualReviewRequired bool              `json:"compliance_manual_review_required"`
}

type SubmitTransactionResponse struct {
	TransactionID      string                `json:"transaction_id"`
	SwitchReference    string                `json:"switch_reference"`
	State              core.TransactionState `json:"state"`
	IDempotentReplay   bool                  `json:"idempotent_replay"`
	SelectedRouteID    core.RouteID          `json:"selected_route_id,omitempty"`
	SelectedProviderID core.ProviderID       `json:"selected_provider_id,omitempty"`
	RouteDecision      core.RouteDecision    `json:"route_decision"`
}

type TransactionRecord struct {
	TransactionID   string                   `json:"transaction_id"`
	SwitchReference string                   `json:"switch_reference"`
	BankID          string                   `json:"bank_id"`
	IDempotencyKey  string                   `json:"idempotency_key"`
	State           core.TransactionState    `json:"state"`
	Request         SubmitTransactionRequest `json:"request"`
	RouteDecision   core.RouteDecision       `json:"route_decision"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
}

type TransactionStore interface {
	Create(record TransactionRecord) TransactionRecord
	GetByIDempotencyKey(bankID, key string) (TransactionRecord, bool)
	GetByTransactionID(transactionID string) (TransactionRecord, bool)
}

type EventSink interface {
	Emit(event core.TransactionEvent)
}

type Service struct {
	mu           sync.Mutex
	registry     core.RouteRegistry
	policy       core.BankPolicy
	health       core.RouteHealthProvider
	transactions TransactionStore
	decisions    core.RouteDecisionStore
	events       EventSink
}

func NewService(registry core.RouteRegistry, policy core.BankPolicy, health core.RouteHealthProvider, transactions TransactionStore, decisions core.RouteDecisionStore, events EventSink) *Service {
	return &Service{
		registry:     registry,
		policy:       policy,
		health:       health,
		transactions: transactions,
		decisions:    decisions,
		events:       events,
	}
}

func (s *Service) Submit(request SubmitTransactionRequest) (SubmitTransactionResponse, error) {
	if err := validate(request); err != nil {
		return SubmitTransactionResponse{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.transactions.GetByIDempotencyKey(request.BankID, request.IDempotencyKey); ok {
		return responseFromRecord(existing, true), nil
	}

	transactionID := NewTransactionID()
	switchReference := "IMSI-" + transactionID
	input := core.EligibilityInput{
		TransactionID:                  transactionID,
		Corridor:                       core.NewCorridor(request.SenderCountry, request.ReceiverCountry),
		PayoutMethod:                   request.PayoutMethod,
		Amount:                         core.NewMoney(request.Currency, request.AmountMinorUnits),
		DestinationBank:                request.DestinationBank,
		ComplianceManualReviewRequired: request.ComplianceManualReviewRequired,
		CircuitBreakerBlockedRoutes:    s.health.BlockedRoutes(),
	}

	decision := core.SelectRoute(s.registry, s.policy, s.health, input)
	state := core.StateAccepted
	if decision.SelectedRoute == nil {
		state = core.StateFailed
	}

	now := time.Now().UTC()
	record := s.transactions.Create(TransactionRecord{
		TransactionID:   transactionID,
		SwitchReference: switchReference,
		BankID:          request.BankID,
		IDempotencyKey:  request.IDempotencyKey,
		State:           state,
		Request:         request,
		RouteDecision:   decision,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	s.decisions.Record(decision)
	s.events.Emit(core.NewTransactionEvent(transactionID+"-created", transactionID, core.StateCreated, core.EventSourceBank, "transaction received from bank channel"))
	s.events.Emit(core.NewTransactionEvent(transactionID+"-accepted", transactionID, state, core.EventSourceSwitch, "transaction intake processed and route decision recorded"))

	return responseFromRecord(record, false), nil
}

func validate(request SubmitTransactionRequest) error {
	if request.BankID == "" {
		return ErrMissingBankID
	}
	if request.IDempotencyKey == "" {
		return ErrMissingIDempotencyKey
	}
	if request.SenderCountry == "" || request.ReceiverCountry == "" {
		return ErrMissingCountries
	}
	if request.AmountMinorUnits == 0 {
		return ErrInvalidAmount
	}
	if request.Currency == "" {
		return ErrMissingCurrency
	}
	if request.PayoutMethod == "" {
		return ErrMissingPayoutMethod
	}
	return nil
}

func responseFromRecord(record TransactionRecord, replay bool) SubmitTransactionResponse {
	response := SubmitTransactionResponse{
		TransactionID:    record.TransactionID,
		SwitchReference:  record.SwitchReference,
		State:            record.State,
		IDempotentReplay: replay,
		RouteDecision:    record.RouteDecision,
	}
	if record.RouteDecision.SelectedRoute != nil {
		response.SelectedRouteID = record.RouteDecision.SelectedRoute.RouteID
		response.SelectedProviderID = record.RouteDecision.SelectedRoute.ProviderID
	}
	return response
}

type InMemoryTransactionStore struct {
	mu            sync.RWMutex
	byID          map[string]TransactionRecord
	byIdempotency map[string]string
}

func NewInMemoryTransactionStore() *InMemoryTransactionStore {
	return &InMemoryTransactionStore{
		byID:          map[string]TransactionRecord{},
		byIdempotency: map[string]string{},
	}
}

func (s *InMemoryTransactionStore) Create(record TransactionRecord) TransactionRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[record.TransactionID] = record
	s.byIdempotency[idempotencyLookupKey(record.BankID, record.IDempotencyKey)] = record.TransactionID
	return record
}

func (s *InMemoryTransactionStore) GetByIDempotencyKey(bankID, key string) (TransactionRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	transactionID, ok := s.byIdempotency[idempotencyLookupKey(bankID, key)]
	if !ok {
		return TransactionRecord{}, false
	}
	record, ok := s.byID[transactionID]
	return record, ok
}

func (s *InMemoryTransactionStore) GetByTransactionID(transactionID string) (TransactionRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.byID[transactionID]
	return record, ok
}

func idempotencyLookupKey(bankID, key string) string {
	return bankID + ":" + key
}

type InMemoryEventSink struct {
	mu     sync.RWMutex
	events []core.TransactionEvent
}

func NewInMemoryEventSink() *InMemoryEventSink {
	return &InMemoryEventSink{events: []core.TransactionEvent{}}
}

func (s *InMemoryEventSink) Emit(event core.TransactionEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

func (s *InMemoryEventSink) Events() []core.TransactionEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	events := make([]core.TransactionEvent, len(s.events))
	copy(events, s.events)
	return events
}

var transactionIDCounter = struct {
	sync.Mutex
	next uint64
}{next: 1}

func NewTransactionID() string {
	transactionIDCounter.Lock()
	defer transactionIDCounter.Unlock()
	id := transactionIDCounter.next
	transactionIDCounter.next++
	return fmt.Sprintf("txn_%012d", id)
}
