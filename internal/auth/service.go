package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrLockedOut          = errors.New("too many failed login attempts")
	ErrNoMappedRole       = errors.New("authenticated identity has no mapped imsi-rails role")
	ErrUnauthorizedDomain = errors.New("email domain is not allowed")
	ErrSelfApproval       = errors.New("maker-checker approval requires a different approver")
)

type Service struct {
	store         Store
	hasher        PasswordHasher
	ldapDirectory LDAPDirectory
	oidc          OIDCAuthenticator
	sessionTTL    time.Duration
	cookieName    string
	secureCookies bool
	throttle      *LoginThrottle
}

type Config struct {
	SessionTTL    time.Duration
	CookieName    string
	SecureCookies bool
}

func NewService(store Store, hasher PasswordHasher, ldapDirectory LDAPDirectory, oidc OIDCAuthenticator, config Config) *Service {
	if config.SessionTTL == 0 {
		config.SessionTTL = 12 * time.Hour
	}
	if config.CookieName == "" {
		config.CookieName = "imsi_session"
	}
	if ldapDirectory == nil {
		ldapDirectory = NewLDAPClient()
	}
	if oidc == nil {
		oidc = NewGoogleOIDCAuthenticator()
	}
	return &Service{
		store:         store,
		hasher:        hasher,
		ldapDirectory: ldapDirectory,
		oidc:          oidc,
		sessionTTL:    config.SessionTTL,
		cookieName:    config.CookieName,
		secureCookies: config.SecureCookies,
		throttle:      NewLoginThrottle(5, 15*time.Minute),
	}
}

func NewDefaultService() (*Service, error) {
	store := NewInMemoryStore()
	service := NewService(store, DefaultPasswordHasher(), nil, nil, Config{})
	if err := service.SeedLocalUser("bank-demo", "admin", "admin@imsi.local", "Operations Admin", "admin123", []Role{RolePlatformAdmin}); err != nil {
		return nil, err
	}
	store.UpsertProvider(AuthProviderConfig{
		ID:           "local:bank-demo",
		BankID:       "bank-demo",
		Type:         ProviderLocal,
		Name:         "Local password login",
		Enabled:      true,
		DefaultRoles: []Role{RoleViewer},
	})
	store.UpsertProvider(AuthProviderConfig{
		ID:      "ldap:bank-demo",
		BankID:  "bank-demo",
		Type:    ProviderLDAP,
		Name:    "Bank LDAP / Active Directory",
		Enabled: false,
		LDAP: &LDAPConfig{
			Host:                 "ldap.bank.example",
			Port:                 636,
			TLSMode:              "tls",
			BaseDN:               "dc=bank,dc=example",
			UserFilter:           "(&(objectClass=user)(sAMAccountName={{username}}))",
			GroupFilter:          "(member={{dn}})",
			UsernameAttribute:    "sAMAccountName",
			EmailAttribute:       "mail",
			DisplayNameAttribute: "displayName",
			GroupNameAttribute:   "cn",
		},
		GroupMapping: []LDAPGroupMapping{
			{ExternalGroup: "IMSI Platform Admins", Roles: []Role{RolePlatformAdmin}},
			{ExternalGroup: "IMSI Operations", Roles: []Role{RoleOpsAnalyst}},
			{ExternalGroup: "IMSI Auditors", Roles: []Role{RoleAuditor}},
		},
		DefaultRoles: []Role{RoleViewer},
	})
	store.UpsertProvider(AuthProviderConfig{
		ID:      "google_oidc:bank-demo",
		BankID:  "bank-demo",
		Type:    ProviderGoogleOIDC,
		Name:    "Google SSO",
		Enabled: false,
		GoogleOIDC: &GoogleOIDCConfig{
			Issuer:         "https://accounts.google.com",
			AllowedDomains: []string{"imsi.local"},
		},
		DefaultRoles: []Role{RoleViewer},
	})
	return service, nil
}

func (s *Service) SeedLocalUser(bankID, username, email, displayName, password string, roles []Role) error {
	user, err := s.store.CreateUser(User{
		BankID:       bankID,
		Username:     username,
		Email:        email,
		DisplayName:  displayName,
		Roles:        roles,
		Status:       UserActive,
		AuthProvider: string(ProviderLocal),
	})
	if err != nil {
		return err
	}
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return err
	}
	return s.store.CreateLocalCredential(LocalCredential{
		BankID:       bankID,
		Username:     username,
		UserID:       user.ID,
		PasswordHash: passwordHash,
	})
}

func (s *Service) LoginLocal(bankID, username, password string) (Session, error) {
	key := bankID + ":" + strings.ToLower(strings.TrimSpace(username))
	if !s.throttle.Allow(key) {
		return Session{}, ErrLockedOut
	}
	credential, ok := s.store.GetLocalCredential(bankID, username)
	if !ok {
		s.throttle.RecordFailure(key)
		return Session{}, ErrInvalidCredentials
	}
	ok, err := s.hasher.Verify(password, credential.PasswordHash)
	if err != nil || !ok {
		s.throttle.RecordFailure(key)
		return Session{}, ErrInvalidCredentials
	}
	user, ok := s.store.GetUser(credential.UserID)
	if !ok {
		s.throttle.RecordFailure(key)
		return Session{}, ErrUserNotFound
	}
	if user.Status != UserActive {
		return Session{}, ErrUserDisabled
	}
	s.throttle.RecordSuccess(key)
	return s.createSession(user)
}

func (s *Service) LoginLDAP(ctx context.Context, bankID, username, password string) (Session, error) {
	provider, ok := s.store.GetProvider(bankID, ProviderLDAP)
	if !ok {
		return Session{}, ErrProviderNotFound
	}
	if !provider.Enabled {
		return Session{}, ErrProviderDisabled
	}
	profile, err := s.ldapDirectory.Authenticate(ctx, provider, username, password)
	if err != nil {
		return Session{}, err
	}
	roles := rolesFromGroups(profile.Groups, provider.GroupMapping)
	if len(roles) == 0 {
		roles = append([]Role(nil), provider.DefaultRoles...)
	}
	if len(roles) == 0 {
		return Session{}, ErrNoMappedRole
	}
	user := s.store.UpsertUser(User{
		BankID:       bankID,
		Username:     profile.Username,
		Email:        profile.Email,
		DisplayName:  profile.DisplayName,
		Roles:        roles,
		Status:       UserActive,
		AuthProvider: provider.ID,
	})
	return s.createSession(user)
}

func (s *Service) StartGoogle(bankID, returnTo string) (OIDCRedirect, error) {
	provider, ok := s.store.GetProvider(bankID, ProviderGoogleOIDC)
	if !ok {
		return OIDCRedirect{}, ErrProviderNotFound
	}
	if !provider.Enabled {
		return OIDCRedirect{}, ErrProviderDisabled
	}
	return s.oidc.Start(context.Background(), provider, returnTo)
}

func (s *Service) CompleteGoogle(ctx context.Context, bankID, state, code string) (Session, error) {
	provider, ok := s.store.GetProvider(bankID, ProviderGoogleOIDC)
	if !ok {
		return Session{}, ErrProviderNotFound
	}
	if !provider.Enabled {
		return Session{}, ErrProviderDisabled
	}
	profile, err := s.oidc.Complete(ctx, provider, state, code)
	if err != nil {
		return Session{}, err
	}
	if !emailDomainAllowed(profile.Email, provider.GoogleOIDC.AllowedDomains) {
		return Session{}, ErrUnauthorizedDomain
	}
	roles := append([]Role(nil), provider.DefaultRoles...)
	if len(roles) == 0 {
		return Session{}, ErrNoMappedRole
	}
	user := s.store.UpsertUser(User{
		BankID:       bankID,
		Username:     profile.Email,
		Email:        profile.Email,
		DisplayName:  profile.DisplayName,
		Roles:        roles,
		Status:       UserActive,
		AuthProvider: provider.ID,
	})
	return s.createSession(user)
}

func (s *Service) createSession(user User) (Session, error) {
	now := time.Now().UTC()
	s.store.RecordLogin(user.ID, now)
	return s.store.CreateSession(user.SessionUser(), s.sessionTTL)
}

func (s *Service) SessionFromRequest(r *http.Request) (Session, bool) {
	cookie, err := r.Cookie(s.cookieName)
	if err != nil || cookie.Value == "" {
		return Session{}, false
	}
	return s.store.GetSession(cookie.Value)
}

func (s *Service) DeleteSessionFromRequest(r *http.Request) {
	cookie, err := r.Cookie(s.cookieName)
	if err == nil {
		s.store.DeleteSession(cookie.Value)
	}
}

func (s *Service) SessionCookie(session Session) *http.Cookie {
	return &http.Cookie{
		Name:     s.cookieName,
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   s.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}
}

func (s *Service) ExpiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     s.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}
}

func (s *Service) ListUsers(bankID string) []User {
	return s.store.ListUsers(bankID)
}

func (s *Service) ListProviders(bankID string) []AuthProviderConfig {
	return s.store.ListProviders(bankID)
}

func (s *Service) UpsertProvider(provider AuthProviderConfig) AuthProviderConfig {
	return s.store.UpsertProvider(provider)
}

func (s *Service) Roles() []map[string]any {
	roles := make([]map[string]any, 0)
	for _, role := range KnownRoles() {
		roles = append(roles, map[string]any{
			"role":        role,
			"permissions": RolePermissions[role],
		})
	}
	return roles
}

func (s *Service) ApprovePolicyDraft(drafterUserID, approverUserID string) error {
	if drafterUserID == approverUserID {
		return ErrSelfApproval
	}
	approver, ok := s.store.GetUser(approverUserID)
	if !ok {
		return ErrUserNotFound
	}
	if !approver.SessionUser().HasPermission(PermissionPolicyApprove) {
		return ErrForbidden
	}
	return nil
}

func rolesFromGroups(groups []string, mappings []LDAPGroupMapping) []Role {
	seen := map[Role]bool{}
	roles := make([]Role, 0)
	for _, group := range groups {
		normalizedGroup := strings.ToLower(strings.TrimSpace(group))
		for _, mapping := range mappings {
			if strings.ToLower(strings.TrimSpace(mapping.ExternalGroup)) != normalizedGroup {
				continue
			}
			for _, role := range mapping.Roles {
				if !seen[role] {
					seen[role] = true
					roles = append(roles, role)
				}
			}
		}
	}
	return roles
}

func emailDomainAllowed(email string, domains []string) bool {
	if len(domains) == 0 {
		return true
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := strings.ToLower(parts[1])
	for _, allowed := range domains {
		if strings.ToLower(strings.TrimSpace(allowed)) == domain {
			return true
		}
	}
	return false
}

type LoginThrottle struct {
	mu          sync.Mutex
	maxAttempts int
	lockFor     time.Duration
	attempts    map[string]loginAttempt
}

type loginAttempt struct {
	Failures    int
	LockedUntil time.Time
}

func NewLoginThrottle(maxAttempts int, lockFor time.Duration) *LoginThrottle {
	return &LoginThrottle{
		maxAttempts: maxAttempts,
		lockFor:     lockFor,
		attempts:    map[string]loginAttempt{},
	}
}

func (t *LoginThrottle) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	attempt := t.attempts[key]
	return attempt.LockedUntil.IsZero() || time.Now().UTC().After(attempt.LockedUntil)
}

func (t *LoginThrottle) RecordFailure(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	attempt := t.attempts[key]
	attempt.Failures++
	if attempt.Failures >= t.maxAttempts {
		attempt.LockedUntil = time.Now().UTC().Add(t.lockFor)
	}
	t.attempts[key] = attempt
}

func (t *LoginThrottle) RecordSuccess(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, key)
}
