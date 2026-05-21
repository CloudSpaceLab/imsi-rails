package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/auth/login", h.login)
	mux.HandleFunc("POST /v1/auth/ldap/login", h.ldapLogin)
	mux.HandleFunc("GET /v1/auth/oidc/google/start", h.googleStart)
	mux.HandleFunc("GET /v1/auth/oidc/google/callback", h.googleCallback)
	mux.HandleFunc("POST /v1/auth/logout", h.logout)
	mux.HandleFunc("GET /v1/auth/me", h.me)

	mux.Handle("GET /v1/admin/users", h.service.Require(PermissionUsersManage, http.HandlerFunc(h.listUsers)))
	mux.Handle("POST /v1/admin/users", h.service.Require(PermissionUsersManage, http.HandlerFunc(h.createUser)))
	mux.Handle("GET /v1/admin/roles", h.service.Require(PermissionUsersManage, http.HandlerFunc(h.listRoles)))
	mux.Handle("POST /v1/admin/roles", h.service.Require(PermissionUsersManage, http.HandlerFunc(h.listRoles)))
	mux.Handle("GET /v1/admin/identity-providers", h.service.Require(PermissionIdentityManage, http.HandlerFunc(h.listProviders)))
	mux.Handle("POST /v1/admin/identity-providers", h.service.Require(PermissionIdentityManage, http.HandlerFunc(h.upsertProvider)))
	mux.Handle("POST /v1/admin/identity-providers/{id}/test", h.service.Require(PermissionIdentityManage, http.HandlerFunc(h.testProvider)))
	mux.Handle("POST /v1/admin/identity-providers/{id}/preview-groups", h.service.Require(PermissionIdentityManage, http.HandlerFunc(h.previewGroups)))
}

type loginRequest struct {
	BankID   string `json:"bank_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	session, err := h.service.LoginLocal(request.BankID, request.Username, request.Password)
	h.writeSessionResponse(w, session, err)
}

func (h *Handler) ldapLogin(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	session, err := h.service.LoginLDAP(r.Context(), request.BankID, request.Username, request.Password)
	h.writeSessionResponse(w, session, err)
}

func (h *Handler) googleStart(w http.ResponseWriter, r *http.Request) {
	bankID := firstNonEmpty(r.URL.Query().Get("bank_id"), "bank-demo")
	redirect, err := h.service.StartGoogle(bankID, r.URL.Query().Get("return_to"))
	if err != nil {
		writeProblem(w, statusForAuthError(err), "oidc_start_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, redirect)
}

func (h *Handler) googleCallback(w http.ResponseWriter, r *http.Request) {
	bankID := firstNonEmpty(r.URL.Query().Get("bank_id"), "bank-demo")
	session, err := h.service.CompleteGoogle(r.Context(), bankID, r.URL.Query().Get("state"), r.URL.Query().Get("code"))
	h.writeSessionResponse(w, session, err)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	h.service.DeleteSessionFromRequest(r)
	http.SetCookie(w, h.service.ExpiredSessionCookie())
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	session, ok := h.service.SessionFromRequest(r)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "A valid session is required.")
		return
	}
	writeJSON(w, http.StatusOK, session.User)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	bankID := r.URL.Query().Get("bank_id")
	writeJSON(w, http.StatusOK, map[string]any{"users": h.service.ListUsers(bankID)})
}

type createUserRequest struct {
	BankID      string `json:"bank_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
	Roles       []Role `json:"roles"`
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	if request.BankID == "" || request.Username == "" || request.Password == "" {
		writeProblem(w, http.StatusBadRequest, "invalid_user", "bank_id, username, and password are required.")
		return
	}
	if len(request.Roles) == 0 {
		request.Roles = []Role{RoleViewer}
	}
	err := h.service.SeedLocalUser(request.BankID, request.Username, request.Email, request.DisplayName, request.Password, request.Roles)
	if err != nil {
		writeProblem(w, statusForAuthError(err), "user_create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *Handler) listRoles(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"roles": h.service.Roles()})
}

func (h *Handler) listProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"identity_providers": h.service.ListProviders(r.URL.Query().Get("bank_id"))})
}

func (h *Handler) upsertProvider(w http.ResponseWriter, r *http.Request) {
	var provider AuthProviderConfig
	if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	if provider.BankID == "" || provider.Type == "" {
		writeProblem(w, http.StatusBadRequest, "invalid_provider", "bank_id and type are required.")
		return
	}
	writeJSON(w, http.StatusOK, h.service.UpsertProvider(provider))
}

func (h *Handler) testProvider(w http.ResponseWriter, r *http.Request) {
	provider, ok := h.service.store.GetProviderByID(strings.TrimSpace(r.PathValue("id")))
	if !ok {
		writeProblem(w, http.StatusNotFound, "provider_not_found", "Identity provider was not found.")
		return
	}
	if provider.Type == ProviderLDAP {
		writeJSON(w, http.StatusOK, ValidateLDAPConfig(provider))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"provider_id": provider.ID, "reachable": provider.Enabled, "warnings": []string{}})
}

func (h *Handler) previewGroups(w http.ResponseWriter, r *http.Request) {
	provider, ok := h.service.store.GetProviderByID(strings.TrimSpace(r.PathValue("id")))
	if !ok {
		writeProblem(w, http.StatusNotFound, "provider_not_found", "Identity provider was not found.")
		return
	}
	groups := make([]map[string]any, 0, len(provider.GroupMapping))
	for _, mapping := range provider.GroupMapping {
		groups = append(groups, map[string]any{
			"external_group": mapping.ExternalGroup,
			"roles":          mapping.Roles,
			"permissions":    PermissionsForRoles(mapping.Roles),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"provider_id": provider.ID, "groups": groups})
}

func (h *Handler) writeSessionResponse(w http.ResponseWriter, session Session, err error) {
	if err != nil {
		writeProblem(w, statusForAuthError(err), "login_failed", err.Error())
		return
	}
	http.SetCookie(w, h.service.SessionCookie(session))
	writeJSON(w, http.StatusOK, session.User)
}

func statusForAuthError(err error) int {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, ErrLockedOut):
		return http.StatusTooManyRequests
	case errors.Is(err, ErrForbidden),
		errors.Is(err, ErrNoMappedRole),
		errors.Is(err, ErrUnauthorizedDomain),
		errors.Is(err, ErrProviderDisabled):
		return http.StatusForbidden
	case errors.Is(err, ErrProviderNotFound),
		errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrDuplicateUser):
		return http.StatusConflict
	default:
		return http.StatusUnprocessableEntity
	}
}
