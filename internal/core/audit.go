package core

import "sync"

type RouteDecisionStore interface {
	Record(decision RouteDecision)
	GetByTransactionID(transactionID string) (RouteDecision, bool)
}

type InMemoryRouteDecisionStore struct {
	mu        sync.RWMutex
	decisions map[string]RouteDecision
}

func NewInMemoryRouteDecisionStore() *InMemoryRouteDecisionStore {
	return &InMemoryRouteDecisionStore{
		decisions: map[string]RouteDecision{},
	}
}

func (s *InMemoryRouteDecisionStore) Record(decision RouteDecision) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.decisions[decision.TransactionID] = decision
}

func (s *InMemoryRouteDecisionStore) GetByTransactionID(transactionID string) (RouteDecision, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	decision, ok := s.decisions[transactionID]
	return decision, ok
}
