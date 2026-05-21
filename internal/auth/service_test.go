package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeLDAP struct {
	profile LDAPProfile
	err     error
}

func (f fakeLDAP) Authenticate(context.Context, AuthProviderConfig, string, string) (LDAPProfile, error) {
	return f.profile, f.err
}

type fakeOIDC struct {
	profile OIDCProfile
	err     error
}

func (f fakeOIDC) Start(context.Context, AuthProviderConfig, string) (OIDCRedirect, error) {
	return OIDCRedirect{URL: "https://accounts.example/auth", State: "state"}, nil
}

func (f fakeOIDC) Complete(context.Context, AuthProviderConfig, string, string) (OIDCProfile, error) {
	return f.profile, f.err
}

func TestLocalLoginAndPasswordHash(t *testing.T) {
	service := testService(t, nil, nil)
	if err := service.SeedLocalUser("bank-a", "ops", "ops@example.com", "Ops User", "correct-password", []Role{RoleOpsLead}); err != nil {
		t.Fatal(err)
	}

	session, err := service.LoginLocal("bank-a", "ops", "correct-password")
	if err != nil {
		t.Fatal(err)
	}
	if !session.User.HasPermission(PermissionPolicyApprove) {
		t.Fatalf("expected ops lead permissions, got %#v", session.User.Permissions)
	}

	if _, err := service.LoginLocal("bank-a", "ops", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestLocalLoginLockout(t *testing.T) {
	service := testService(t, nil, nil)
	service.throttle = NewLoginThrottle(2, 999999999)
	if err := service.SeedLocalUser("bank-a", "ops", "ops@example.com", "Ops User", "correct-password", []Role{RoleOpsLead}); err != nil {
		t.Fatal(err)
	}

	_, _ = service.LoginLocal("bank-a", "ops", "wrong")
	_, _ = service.LoginLocal("bank-a", "ops", "wrong")
	if _, err := service.LoginLocal("bank-a", "ops", "correct-password"); !errors.Is(err, ErrLockedOut) {
		t.Fatalf("expected lockout, got %v", err)
	}
}

func TestLDAPLoginMapsGroupsAndCreatesUser(t *testing.T) {
	ldap := fakeLDAP{profile: LDAPProfile{Username: "a.ope", Email: "a.ope@bank.test", DisplayName: "A Ope", Groups: []string{"IMSI Operations"}}}
	service := testService(t, ldap, nil)
	service.UpsertProvider(AuthProviderConfig{
		ID:      "ldap:bank-a",
		BankID:  "bank-a",
		Type:    ProviderLDAP,
		Enabled: true,
		LDAP:    &LDAPConfig{Host: "ldap.bank.test", TLSMode: "tls"},
		GroupMapping: []LDAPGroupMapping{
			{ExternalGroup: "IMSI Operations", Roles: []Role{RoleOpsAnalyst}},
		},
	})

	session, err := service.LoginLDAP(context.Background(), "bank-a", "a.ope", "password")
	if err != nil {
		t.Fatal(err)
	}
	if !session.User.HasPermission(PermissionTransactionsTrace) {
		t.Fatalf("expected trace permission from LDAP group mapping")
	}
	if _, ok := service.store.GetUserByLogin("bank-a", "a.ope"); !ok {
		t.Fatalf("expected jit user to be created")
	}
}

func TestLDAPLoginRejectsUnmappedGroups(t *testing.T) {
	ldap := fakeLDAP{profile: LDAPProfile{Username: "audit", Email: "audit@bank.test", Groups: []string{"Unknown"}}}
	service := testService(t, ldap, nil)
	service.UpsertProvider(AuthProviderConfig{ID: "ldap:bank-a", BankID: "bank-a", Type: ProviderLDAP, Enabled: true, LDAP: &LDAPConfig{Host: "ldap"}})

	if _, err := service.LoginLDAP(context.Background(), "bank-a", "audit", "password"); !errors.Is(err, ErrNoMappedRole) {
		t.Fatalf("expected unmapped group denial, got %v", err)
	}
}

func TestGoogleOIDCDomainAndMakerChecker(t *testing.T) {
	oidc := fakeOIDC{profile: OIDCProfile{Email: "lead@bank.test", DisplayName: "Lead"}}
	service := testService(t, nil, oidc)
	service.UpsertProvider(AuthProviderConfig{
		ID:           "google_oidc:bank-a",
		BankID:       "bank-a",
		Type:         ProviderGoogleOIDC,
		Enabled:      true,
		GoogleOIDC:   &GoogleOIDCConfig{AllowedDomains: []string{"bank.test"}},
		DefaultRoles: []Role{RoleViewer},
	})
	session, err := service.CompleteGoogle(context.Background(), "bank-a", "state", "code")
	if err != nil {
		t.Fatal(err)
	}
	if session.User.AuthProvider != "google_oidc:bank-a" {
		t.Fatalf("expected google provider, got %s", session.User.AuthProvider)
	}

	approver, err := service.store.CreateUser(User{BankID: "bank-a", Username: "approver", Roles: []Role{RoleOpsLead}, Status: UserActive})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ApprovePolicyDraft(approver.ID, approver.ID); !errors.Is(err, ErrSelfApproval) {
		t.Fatalf("expected self approval rejection, got %v", err)
	}
}

func TestAuthorizationDenyByDefault(t *testing.T) {
	user := User{Roles: []Role{RoleViewer}}.SessionUser()
	if user.HasPermission(PermissionPolicyActivate) {
		t.Fatalf("viewer must not be able to activate policy")
	}
	if !user.HasPermission(PermissionDashboardRead) {
		t.Fatalf("viewer should read dashboard")
	}
}

func testService(t *testing.T, ldap LDAPDirectory, oidc OIDCAuthenticator) *Service {
	t.Helper()
	return NewService(NewInMemoryStore(), PasswordHasher{MemoryKiB: 1024, Iterations: 1, Parallelism: 1, SaltBytes: 8, KeyBytes: 16}, ldap, oidc, Config{})
}
