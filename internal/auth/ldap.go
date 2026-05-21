package auth

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
)

var (
	ErrLDAPConfigMissing = errors.New("ldap configuration is missing")
	ErrLDAPUserNotFound  = errors.New("ldap user was not found")
)

type LDAPProfile struct {
	DN          string   `json:"dn"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	Groups      []string `json:"groups"`
}

type LDAPDirectory interface {
	Authenticate(ctx context.Context, provider AuthProviderConfig, username, password string) (LDAPProfile, error)
}

type LDAPClient struct {
	DialTimeout time.Duration
}

func NewLDAPClient() *LDAPClient {
	return &LDAPClient{DialTimeout: 5 * time.Second}
}

func (c *LDAPClient) Authenticate(ctx context.Context, provider AuthProviderConfig, username, password string) (LDAPProfile, error) {
	if provider.LDAP == nil {
		return LDAPProfile{}, ErrLDAPConfigMissing
	}
	config := *provider.LDAP
	conn, err := c.dial(ctx, config)
	if err != nil {
		return LDAPProfile{}, err
	}
	defer conn.Close()

	if config.BindDN != "" {
		if err := conn.Bind(config.BindDN, ""); err != nil {
			return LDAPProfile{}, err
		}
	}

	userEntry, err := searchSingle(conn, config.BaseDN, renderLDAPFilter(config.UserFilter, map[string]string{"username": username}), []string{
		config.UsernameAttribute,
		config.EmailAttribute,
		config.DisplayNameAttribute,
		"dn",
	})
	if err != nil {
		return LDAPProfile{}, err
	}
	if userEntry == nil {
		return LDAPProfile{}, ErrLDAPUserNotFound
	}
	if err := conn.Bind(userEntry.DN, password); err != nil {
		return LDAPProfile{}, ErrInvalidCredentials
	}

	groupEntries, err := searchMany(conn, config.BaseDN, renderLDAPFilter(config.GroupFilter, map[string]string{"dn": userEntry.DN, "username": username}), []string{config.GroupNameAttribute})
	if err != nil {
		return LDAPProfile{}, err
	}
	groups := make([]string, 0, len(groupEntries))
	for _, entry := range groupEntries {
		if group := entry.GetAttributeValue(config.GroupNameAttribute); group != "" {
			groups = append(groups, group)
		}
	}

	return LDAPProfile{
		DN:          userEntry.DN,
		Username:    firstNonEmpty(userEntry.GetAttributeValue(config.UsernameAttribute), username),
		Email:       userEntry.GetAttributeValue(config.EmailAttribute),
		DisplayName: firstNonEmpty(userEntry.GetAttributeValue(config.DisplayNameAttribute), username),
		Groups:      groups,
	}, nil
}

func (c *LDAPClient) dial(ctx context.Context, config LDAPConfig) (*ldap.Conn, error) {
	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	dialer := &net.Dialer{Timeout: c.DialTimeout}
	tlsConfig := &tls.Config{ServerName: config.Host, MinVersion: tls.VersionTLS12}

	switch strings.ToLower(config.TLSMode) {
	case "tls":
		raw, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
		if err != nil {
			return nil, err
		}
		return ldap.NewConn(raw, true), nil
	default:
		raw, err := dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			return nil, err
		}
		conn := ldap.NewConn(raw, false)
		conn.Start()
		if strings.EqualFold(config.TLSMode, "starttls") {
			if err := conn.StartTLS(tlsConfig); err != nil {
				conn.Close()
				return nil, err
			}
		}
		return conn, nil
	}
}

func searchSingle(conn *ldap.Conn, baseDN, filter string, attributes []string) (*ldap.Entry, error) {
	entries, err := searchMany(conn, baseDN, filter, attributes)
	if err != nil || len(entries) == 0 {
		return nil, err
	}
	return entries[0], nil
}

func searchMany(conn *ldap.Conn, baseDN, filter string, attributes []string) ([]*ldap.Entry, error) {
	request := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		20,
		10,
		false,
		filter,
		attributes,
		nil,
	)
	response, err := conn.Search(request)
	if err != nil {
		return nil, err
	}
	return response.Entries, nil
}

func renderLDAPFilter(template string, values map[string]string) string {
	output := template
	for key, value := range values {
		output = strings.ReplaceAll(output, "{{"+key+"}}", ldap.EscapeFilter(value))
	}
	return output
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

type LDAPTestResult struct {
	ProviderID string   `json:"provider_id"`
	Reachable  bool     `json:"reachable"`
	Warnings   []string `json:"warnings"`
}

func ValidateLDAPConfig(provider AuthProviderConfig) LDAPTestResult {
	result := LDAPTestResult{ProviderID: provider.ID, Reachable: provider.Enabled}
	if provider.LDAP == nil {
		result.Warnings = append(result.Warnings, "ldap configuration is missing")
		return result
	}
	if provider.LDAP.TLSMode == "none" {
		result.Warnings = append(result.Warnings, "plain LDAP should be used only in controlled development networks")
	}
	for _, pair := range []struct {
		label string
		value string
	}{
		{"host", provider.LDAP.Host},
		{"base_dn", provider.LDAP.BaseDN},
		{"user_filter", provider.LDAP.UserFilter},
		{"group_filter", provider.LDAP.GroupFilter},
	} {
		if strings.TrimSpace(pair.value) == "" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s is required", pair.label))
		}
	}
	return result
}
