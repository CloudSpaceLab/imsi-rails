package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user was not found")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrDuplicateUser      = errors.New("user already exists")
	ErrProviderNotFound   = errors.New("identity provider was not found")
	ErrProviderDisabled   = errors.New("identity provider is disabled")
	ErrSessionNotFound    = errors.New("session was not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type LocalCredential struct {
	BankID       string    `json:"bank_id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	UserID       string    `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	Token     string      `json:"-"`
	User      SessionUser `json:"user"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at"`
}

type Store interface {
	CreateUser(user User) (User, error)
	UpsertUser(user User) User
	GetUser(id string) (User, bool)
	GetUserByLogin(bankID, username string) (User, bool)
	ListUsers(bankID string) []User
	RecordLogin(userID string, at time.Time)
	CreateLocalCredential(credential LocalCredential) error
	GetLocalCredential(bankID, username string) (LocalCredential, bool)
	UpsertProvider(provider AuthProviderConfig) AuthProviderConfig
	GetProvider(bankID string, providerType ProviderType) (AuthProviderConfig, bool)
	GetProviderByID(id string) (AuthProviderConfig, bool)
	ListProviders(bankID string) []AuthProviderConfig
	CreateSession(user SessionUser, ttl time.Duration) (Session, error)
	GetSession(token string) (Session, bool)
	DeleteSession(token string)
}

type InMemoryStore struct {
	mu               sync.RWMutex
	usersByID        map[string]User
	usersByLogin     map[string]string
	localCredentials map[string]LocalCredential
	providersByID    map[string]AuthProviderConfig
	providersByBank  map[string][]string
	sessions         map[string]Session
	nextUserID       uint64
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		usersByID:        map[string]User{},
		usersByLogin:     map[string]string{},
		localCredentials: map[string]LocalCredential{},
		providersByID:    map[string]AuthProviderConfig{},
		providersByBank:  map[string][]string{},
		sessions:         map[string]Session{},
		nextUserID:       1,
	}
}

func (s *InMemoryStore) CreateUser(user User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := loginKey(user.BankID, user.Username)
	if _, exists := s.usersByLogin[key]; exists {
		return User{}, ErrDuplicateUser
	}
	if user.ID == "" {
		user.ID = s.newUserIDLocked()
	}
	if user.Status == "" {
		user.Status = UserActive
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	s.usersByID[user.ID] = user
	s.usersByLogin[key] = user.ID
	if user.Email != "" {
		s.usersByLogin[loginKey(user.BankID, user.Email)] = user.ID
	}
	return user, nil
}

func (s *InMemoryStore) UpsertUser(user User) User {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingID, ok := s.usersByLogin[loginKey(user.BankID, user.Username)]; ok && user.ID == "" {
		user.ID = existingID
	}
	if user.ID == "" {
		user.ID = s.newUserIDLocked()
	}
	existing, ok := s.usersByID[user.ID]
	if ok && user.CreatedAt.IsZero() {
		user.CreatedAt = existing.CreatedAt
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	if user.Status == "" {
		user.Status = UserActive
	}
	s.usersByID[user.ID] = user
	s.usersByLogin[loginKey(user.BankID, user.Username)] = user.ID
	if user.Email != "" {
		s.usersByLogin[loginKey(user.BankID, user.Email)] = user.ID
	}
	return user
}

func (s *InMemoryStore) GetUser(id string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.usersByID[id]
	return user, ok
}

func (s *InMemoryStore) GetUserByLogin(bankID, username string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userID, ok := s.usersByLogin[loginKey(bankID, username)]
	if !ok {
		return User{}, false
	}
	user, ok := s.usersByID[userID]
	return user, ok
}

func (s *InMemoryStore) ListUsers(bankID string) []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]User, 0)
	for _, user := range s.usersByID {
		if bankID == "" || user.BankID == bankID {
			users = append(users, user)
		}
	}
	return users
}

func (s *InMemoryStore) RecordLogin(userID string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.usersByID[userID]
	if !ok {
		return
	}
	user.LastLoginAt = &at
	s.usersByID[user.ID] = user
}

func (s *InMemoryStore) CreateLocalCredential(credential LocalCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := loginKey(credential.BankID, credential.Username)
	if _, exists := s.localCredentials[key]; exists {
		return ErrDuplicateUser
	}
	if credential.CreatedAt.IsZero() {
		credential.CreatedAt = time.Now().UTC()
	}
	s.localCredentials[key] = credential
	return nil
}

func (s *InMemoryStore) GetLocalCredential(bankID, username string) (LocalCredential, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	credential, ok := s.localCredentials[loginKey(bankID, username)]
	return credential, ok
}

func (s *InMemoryStore) UpsertProvider(provider AuthProviderConfig) AuthProviderConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if provider.CreatedAt.IsZero() {
		provider.CreatedAt = now
	}
	provider.UpdatedAt = now
	if provider.ID == "" {
		provider.ID = string(provider.Type) + ":" + provider.BankID
	}
	if _, exists := s.providersByID[provider.ID]; !exists {
		s.providersByBank[provider.BankID] = append(s.providersByBank[provider.BankID], provider.ID)
	}
	s.providersByID[provider.ID] = provider
	return provider
}

func (s *InMemoryStore) GetProvider(bankID string, providerType ProviderType) (AuthProviderConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, id := range s.providersByBank[bankID] {
		provider := s.providersByID[id]
		if provider.Type == providerType {
			return provider, true
		}
	}
	return AuthProviderConfig{}, false
}

func (s *InMemoryStore) GetProviderByID(id string) (AuthProviderConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	provider, ok := s.providersByID[id]
	return provider, ok
}

func (s *InMemoryStore) ListProviders(bankID string) []AuthProviderConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	providers := make([]AuthProviderConfig, 0)
	for _, provider := range s.providersByID {
		if bankID == "" || provider.BankID == bankID {
			providers = append(providers, provider)
		}
	}
	return providers
}

func (s *InMemoryStore) CreateSession(user SessionUser, ttl time.Duration) (Session, error) {
	token, err := randomToken(32)
	if err != nil {
		return Session{}, err
	}
	now := time.Now().UTC()
	session := Session{
		Token:     token,
		User:      user,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = session
	return session, nil
}

func (s *InMemoryStore) GetSession(token string) (Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[token]
	if !ok || time.Now().UTC().After(session.ExpiresAt) {
		return Session{}, false
	}
	return session, true
}

func (s *InMemoryStore) DeleteSession(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func (s *InMemoryStore) newUserIDLocked() string {
	id := s.nextUserID
	s.nextUserID++
	return "usr_" + strconvBase36(id)
}

func loginKey(bankID, username string) string {
	return strings.ToLower(strings.TrimSpace(bankID)) + ":" + strings.ToLower(strings.TrimSpace(username))
}

func randomToken(byteCount int) (string, error) {
	value := make([]byte, byteCount)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func strconvBase36(value uint64) string {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	if value == 0 {
		return "0"
	}
	out := make([]byte, 0, 13)
	for value > 0 {
		out = append([]byte{alphabet[value%36]}, out...)
		value /= 36
	}
	return string(out)
}
