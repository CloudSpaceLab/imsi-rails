package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/CloudSpaceLab/imsi-rails/internal/auth"
	"github.com/CloudSpaceLab/imsi-rails/internal/core"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/health/samples", h.ingestSample)
	mux.HandleFunc("GET /v1/health/routes/{route_id}", h.getRouteHealth)
}

func (h *Handler) RegisterProtected(mux *http.ServeMux, require auth.Authorizer) {
	mux.Handle("POST /v1/health/samples", require(auth.PermissionProvidersManage, http.HandlerFunc(h.ingestSample)))
	mux.Handle("GET /v1/health/routes/{route_id}", require(auth.PermissionDashboardRead, http.HandlerFunc(h.getRouteHealth)))
}

func (h *Handler) ingestSample(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request IngestSampleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	response, err := h.service.Ingest(request)
	if err != nil {
		writeProblem(w, statusForError(err), "invalid_health_sample", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) getRouteHealth(w http.ResponseWriter, r *http.Request) {
	routeID := core.RouteID(strings.TrimSpace(r.PathValue("route_id")))
	if routeID == "" {
		writeProblem(w, http.StatusBadRequest, "missing_route_id", "route_id is required.")
		return
	}

	response, ok := h.service.LatestRoute(routeID)
	if !ok {
		writeProblem(w, http.StatusNotFound, "route_health_not_found", "Route health was not found.")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, ErrMissingProviderID),
		errors.Is(err, ErrMissingSignalType),
		errors.Is(err, ErrInvalidSignalPayload),
		errors.Is(err, ErrInvalidProviderStatus),
		errors.Is(err, ErrInvalidRate),
		errors.Is(err, ErrInvalidTransaction):
		return http.StatusBadRequest
	default:
		return http.StatusUnprocessableEntity
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeProblem(w http.ResponseWriter, status int, code, detail string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":   code,
			"detail": detail,
		},
	})
}
