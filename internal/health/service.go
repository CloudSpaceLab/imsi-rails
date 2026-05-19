package health

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/CloudSpaceLab/imsi-rails/internal/core"
)

type SignalType string

const (
	SignalProviderAPIStatus  SignalType = "provider_api_status"
	SignalTimeoutRate        SignalType = "timeout_rate"
	SignalErrorRate          SignalType = "error_rate"
	SignalCallbackLag        SignalType = "callback_lag"
	SignalTransactionOutcome SignalType = "transaction_outcome"
)

type State string

const (
	StateUnknown  State = "unknown"
	StateHealthy  State = "healthy"
	StateWatch    State = "watch"
	StateDegraded State = "degraded"
)

type ProviderAPIStatus string

const (
	ProviderAPIUp       ProviderAPIStatus = "up"
	ProviderAPIDegraded ProviderAPIStatus = "degraded"
	ProviderAPIDown     ProviderAPIStatus = "down"
)

var (
	ErrMissingProviderID     = errors.New("provider_id is required")
	ErrMissingSignalType     = errors.New("signal_type is required")
	ErrInvalidSignalPayload  = errors.New("exactly one signal payload matching signal_type is required")
	ErrInvalidProviderStatus = errors.New("provider_api.status must be up, degraded, or down")
	ErrInvalidRate           = errors.New("rate signal requires rate_bps or a valid affected_count/total_count pair")
	ErrInvalidTransaction    = errors.New("transaction_outcome requires transaction_id and state")
)

type IngestSampleRequest struct {
	ProviderID         core.ProviderID     `json:"provider_id"`
	RouteID            core.RouteID        `json:"route_id,omitempty"`
	Corridor           *core.Corridor      `json:"corridor,omitempty"`
	PayoutMethod       core.PayoutMethod   `json:"payout_method,omitempty"`
	SignalType         SignalType          `json:"signal_type"`
	ObservedAt         time.Time           `json:"observed_at,omitempty"`
	WindowSeconds      uint32              `json:"window_seconds,omitempty"`
	ProviderAPI        *ProviderAPISignal  `json:"provider_api,omitempty"`
	TimeoutRate        *RateSignal         `json:"timeout_rate,omitempty"`
	ErrorRate          *RateSignal         `json:"error_rate,omitempty"`
	CallbackLag        *CallbackLagSignal  `json:"callback_lag,omitempty"`
	TransactionOutcome *TransactionOutcome `json:"transaction_outcome,omitempty"`
}

type ProviderAPISignal struct {
	Status         ProviderAPIStatus `json:"status"`
	HTTPStatusCode int               `json:"http_status_code,omitempty"`
	LatencyMS      uint64            `json:"latency_ms,omitempty"`
}

type RateSignal struct {
	RateBps       *uint16 `json:"rate_bps,omitempty"`
	AffectedCount uint64  `json:"affected_count,omitempty"`
	TotalCount    uint64  `json:"total_count,omitempty"`
}

type CallbackLagSignal struct {
	P95LagMS         uint64 `json:"p95_lag_ms"`
	MaxLagMS         uint64 `json:"max_lag_ms,omitempty"`
	DelayedCallbacks uint64 `json:"delayed_callbacks,omitempty"`
	TotalCallbacks   uint64 `json:"total_callbacks,omitempty"`
}

type TransactionOutcome struct {
	TransactionID     string                `json:"transaction_id"`
	State             core.TransactionState `json:"state"`
	TimeToCreditMS    uint64                `json:"time_to_credit_ms,omitempty"`
	ProviderReference string                `json:"provider_reference,omitempty"`
}

type HealthSample struct {
	SampleID           string                    `json:"sample_id"`
	ProviderID         core.ProviderID           `json:"provider_id"`
	RouteID            core.RouteID              `json:"route_id,omitempty"`
	Corridor           *core.Corridor            `json:"corridor,omitempty"`
	PayoutMethod       core.PayoutMethod         `json:"payout_method,omitempty"`
	SignalType         SignalType                `json:"signal_type"`
	ObservedAt         time.Time                 `json:"observed_at"`
	RecordedAt         time.Time                 `json:"recorded_at"`
	WindowSeconds      uint32                    `json:"window_seconds,omitempty"`
	State              State                     `json:"state"`
	ProviderAPI        *ProviderAPISignal        `json:"provider_api,omitempty"`
	TimeoutRate        *RateSignal               `json:"timeout_rate,omitempty"`
	ErrorRate          *RateSignal               `json:"error_rate,omitempty"`
	CallbackLag        *CallbackLagSignal        `json:"callback_lag,omitempty"`
	TransactionOutcome *TransactionOutcome       `json:"transaction_outcome,omitempty"`
	RouteSnapshot      *core.RouteHealthSnapshot `json:"route_snapshot,omitempty"`
}

type HealthStateChangeEvent struct {
	EventID       string          `json:"event_id"`
	SampleID      string          `json:"sample_id"`
	ProviderID    core.ProviderID `json:"provider_id"`
	RouteID       core.RouteID    `json:"route_id,omitempty"`
	PreviousState State           `json:"previous_state"`
	CurrentState  State           `json:"current_state"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Reason        SignalType      `json:"reason"`
}

type IngestSampleResponse struct {
	Sample      HealthSample            `json:"sample"`
	StateChange *HealthStateChangeEvent `json:"state_change,omitempty"`
}

type RouteHealthResponse struct {
	RouteID  core.RouteID             `json:"route_id"`
	State    State                    `json:"state"`
	Snapshot core.RouteHealthSnapshot `json:"snapshot"`
}

type Store interface {
	Record(sample HealthSample) (previous State, changed bool)
	LatestRoute(routeID core.RouteID) (core.RouteHealthSnapshot, State, bool)
	RouteHealthBook() core.RouteHealthBook
}

type EventSink interface {
	Emit(event HealthStateChangeEvent)
}

type Service struct {
	store  Store
	events EventSink
}

func NewService(store Store, events EventSink) *Service {
	return &Service{store: store, events: events}
}

func (s *Service) Ingest(request IngestSampleRequest) (IngestSampleResponse, error) {
	if err := validate(request); err != nil {
		return IngestSampleResponse{}, err
	}

	observedAt := request.ObservedAt
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}

	state, err := stateFromRequest(request)
	if err != nil {
		return IngestSampleResponse{}, err
	}

	var snapshot *core.RouteHealthSnapshot
	if request.RouteID != "" {
		previous, _, ok := s.store.LatestRoute(request.RouteID)
		if !ok {
			previous = core.DefaultRouteHealthSnapshot()
		}
		next := updateSnapshot(previous, request, observedAt)
		snapshot = &next
	}

	recordedAt := time.Now().UTC()
	sample := HealthSample{
		SampleID:           NewSampleID(),
		ProviderID:         request.ProviderID,
		RouteID:            request.RouteID,
		Corridor:           request.Corridor,
		PayoutMethod:       request.PayoutMethod,
		SignalType:         request.SignalType,
		ObservedAt:         observedAt,
		RecordedAt:         recordedAt,
		WindowSeconds:      request.WindowSeconds,
		State:              state,
		ProviderAPI:        request.ProviderAPI,
		TimeoutRate:        request.TimeoutRate,
		ErrorRate:          request.ErrorRate,
		CallbackLag:        request.CallbackLag,
		TransactionOutcome: request.TransactionOutcome,
		RouteSnapshot:      snapshot,
	}

	previous, changed := s.store.Record(sample)
	response := IngestSampleResponse{Sample: sample}
	if changed {
		event := HealthStateChangeEvent{
			EventID:       sample.SampleID + "-state-change",
			SampleID:      sample.SampleID,
			ProviderID:    sample.ProviderID,
			RouteID:       sample.RouteID,
			PreviousState: previous,
			CurrentState:  sample.State,
			OccurredAt:    recordedAt,
			Reason:        sample.SignalType,
		}
		s.events.Emit(event)
		response.StateChange = &event
	}

	return response, nil
}

func (s *Service) LatestRoute(routeID core.RouteID) (RouteHealthResponse, bool) {
	snapshot, state, ok := s.store.LatestRoute(routeID)
	if !ok {
		return RouteHealthResponse{}, false
	}
	return RouteHealthResponse{RouteID: routeID, State: state, Snapshot: snapshot}, true
}

func (s *Service) RouteHealthBook() core.RouteHealthBook {
	return s.store.RouteHealthBook()
}

func (s *Service) SnapshotFor(routeID core.RouteID) core.RouteHealthSnapshot {
	response, ok := s.LatestRoute(routeID)
	if !ok {
		return core.DefaultRouteHealthSnapshot()
	}
	return response.Snapshot
}

func validate(request IngestSampleRequest) error {
	if request.ProviderID == "" {
		return ErrMissingProviderID
	}
	if request.SignalType == "" {
		return ErrMissingSignalType
	}

	payloads := 0
	if request.ProviderAPI != nil {
		payloads++
	}
	if request.TimeoutRate != nil {
		payloads++
	}
	if request.ErrorRate != nil {
		payloads++
	}
	if request.CallbackLag != nil {
		payloads++
	}
	if request.TransactionOutcome != nil {
		payloads++
	}
	if payloads != 1 {
		return ErrInvalidSignalPayload
	}

	switch request.SignalType {
	case SignalProviderAPIStatus:
		if request.ProviderAPI == nil {
			return ErrInvalidSignalPayload
		}
		switch request.ProviderAPI.Status {
		case ProviderAPIUp, ProviderAPIDegraded, ProviderAPIDown:
			return nil
		default:
			return ErrInvalidProviderStatus
		}
	case SignalTimeoutRate:
		if request.TimeoutRate == nil {
			return ErrInvalidSignalPayload
		}
		_, err := rateBps(*request.TimeoutRate)
		return err
	case SignalErrorRate:
		if request.ErrorRate == nil {
			return ErrInvalidSignalPayload
		}
		_, err := rateBps(*request.ErrorRate)
		return err
	case SignalCallbackLag:
		if request.CallbackLag == nil {
			return ErrInvalidSignalPayload
		}
		return nil
	case SignalTransactionOutcome:
		if request.TransactionOutcome == nil {
			return ErrInvalidSignalPayload
		}
		if request.TransactionOutcome.TransactionID == "" || request.TransactionOutcome.State == "" {
			return ErrInvalidTransaction
		}
		return nil
	default:
		return ErrInvalidSignalPayload
	}
}

func stateFromRequest(request IngestSampleRequest) (State, error) {
	switch request.SignalType {
	case SignalProviderAPIStatus:
		switch request.ProviderAPI.Status {
		case ProviderAPIUp:
			return StateHealthy, nil
		case ProviderAPIDegraded:
			return StateWatch, nil
		case ProviderAPIDown:
			return StateDegraded, nil
		}
	case SignalTimeoutRate:
		bps, err := rateBps(*request.TimeoutRate)
		if err != nil {
			return StateUnknown, err
		}
		return stateFromRate(bps), nil
	case SignalErrorRate:
		bps, err := rateBps(*request.ErrorRate)
		if err != nil {
			return StateUnknown, err
		}
		return stateFromRate(bps), nil
	case SignalCallbackLag:
		return stateFromCallbackLag(request.CallbackLag.P95LagMS), nil
	case SignalTransactionOutcome:
		return stateFromTransactionOutcome(request.TransactionOutcome.State), nil
	}
	return StateUnknown, ErrInvalidSignalPayload
}

func stateFromRate(rate uint16) State {
	switch {
	case rate >= 1_000:
		return StateDegraded
	case rate >= 300:
		return StateWatch
	default:
		return StateHealthy
	}
}

func stateFromCallbackLag(p95LagMS uint64) State {
	switch {
	case p95LagMS >= 120_000:
		return StateDegraded
	case p95LagMS >= 60_000:
		return StateWatch
	default:
		return StateHealthy
	}
}

func stateFromTransactionOutcome(state core.TransactionState) State {
	if state.ValueDelivered() {
		return StateHealthy
	}
	switch state {
	case core.StateFailed, core.StateReversed, core.StateDisputed:
		return StateDegraded
	default:
		return StateWatch
	}
}

func updateSnapshot(previous core.RouteHealthSnapshot, request IngestSampleRequest, observedAt time.Time) core.RouteHealthSnapshot {
	next := previous
	next.ObservedAt = observedAt

	switch request.SignalType {
	case SignalProviderAPIStatus:
		next.P95LatencyMS = request.ProviderAPI.LatencyMS
		switch request.ProviderAPI.Status {
		case ProviderAPIUp:
			next.SuccessRateBps = 10_000
			next.UptimeBps = 10_000
		case ProviderAPIDegraded:
			next.SuccessRateBps = minBps(next.SuccessRateBps, 8_500)
			next.UptimeBps = minBps(next.UptimeBps, 8_500)
		case ProviderAPIDown:
			next.SuccessRateBps = 0
			next.UptimeBps = 0
		}
	case SignalTimeoutRate:
		if bps, err := rateBps(*request.TimeoutRate); err == nil {
			next.SuccessRateBps = 10_000 - bps
			next.UptimeBps = 10_000 - bps
		}
	case SignalErrorRate:
		if bps, err := rateBps(*request.ErrorRate); err == nil {
			next.SuccessRateBps = 10_000 - bps
		}
	case SignalCallbackLag:
		next.P95LatencyMS = request.CallbackLag.P95LagMS
	case SignalTransactionOutcome:
		if request.TransactionOutcome.State.ValueDelivered() {
			next.SuccessRateBps = 10_000
		} else if stateFromTransactionOutcome(request.TransactionOutcome.State) == StateDegraded {
			next.SuccessRateBps = 0
		}
		if request.TransactionOutcome.TimeToCreditMS > 0 {
			next.P95LatencyMS = request.TransactionOutcome.TimeToCreditMS
		}
	}

	return next
}

func rateBps(signal RateSignal) (uint16, error) {
	if signal.RateBps != nil {
		if *signal.RateBps > 10_000 {
			return 0, ErrInvalidRate
		}
		return *signal.RateBps, nil
	}
	if signal.TotalCount == 0 || signal.AffectedCount > signal.TotalCount {
		return 0, ErrInvalidRate
	}
	return uint16((signal.AffectedCount * 10_000) / signal.TotalCount), nil
}

func minBps(left, right uint16) uint16 {
	if left < right {
		return left
	}
	return right
}

type InMemoryStore struct {
	mu             sync.RWMutex
	samples        []HealthSample
	routeSnapshots map[core.RouteID]core.RouteHealthSnapshot
	states         map[string]State
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		samples:        []HealthSample{},
		routeSnapshots: map[core.RouteID]core.RouteHealthSnapshot{},
		states:         map[string]State{},
	}
}

func (s *InMemoryStore) Record(sample HealthSample) (State, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.samples = append(s.samples, sample)
	if sample.RouteID != "" && sample.RouteSnapshot != nil {
		s.routeSnapshots[sample.RouteID] = *sample.RouteSnapshot
	}

	key := stateKey(sample)
	previous, ok := s.states[key]
	if !ok {
		previous = StateUnknown
	}
	s.states[key] = sample.State
	return previous, previous != sample.State
}

func (s *InMemoryStore) LatestRoute(routeID core.RouteID) (core.RouteHealthSnapshot, State, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, ok := s.routeSnapshots[routeID]
	if !ok {
		return core.RouteHealthSnapshot{}, StateUnknown, false
	}
	state := s.states["route:"+string(routeID)]
	if state == "" {
		state = StateUnknown
	}
	return snapshot, state, true
}

func (s *InMemoryStore) RouteHealthBook() core.RouteHealthBook {
	s.mu.RLock()
	defer s.mu.RUnlock()

	book := core.NewRouteHealthBook()
	for routeID, snapshot := range s.routeSnapshots {
		book.Routes[routeID] = snapshot
	}
	return book
}

func (s *InMemoryStore) Samples() []HealthSample {
	s.mu.RLock()
	defer s.mu.RUnlock()

	samples := make([]HealthSample, len(s.samples))
	copy(samples, s.samples)
	return samples
}

func stateKey(sample HealthSample) string {
	if sample.RouteID != "" {
		return "route:" + string(sample.RouteID)
	}
	return "provider:" + string(sample.ProviderID)
}

type InMemoryEventSink struct {
	mu     sync.RWMutex
	events []HealthStateChangeEvent
}

func NewInMemoryEventSink() *InMemoryEventSink {
	return &InMemoryEventSink{events: []HealthStateChangeEvent{}}
}

func (s *InMemoryEventSink) Emit(event HealthStateChangeEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

func (s *InMemoryEventSink) Events() []HealthStateChangeEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]HealthStateChangeEvent, len(s.events))
	copy(events, s.events)
	return events
}

var sampleIDCounter = struct {
	sync.Mutex
	next uint64
}{next: 1}

func NewSampleID() string {
	sampleIDCounter.Lock()
	defer sampleIDCounter.Unlock()
	id := sampleIDCounter.next
	sampleIDCounter.next++
	return fmt.Sprintf("health_%012d", id)
}
