package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CloudSpaceLab/imsi-rails/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterProtected(mux *http.ServeMux, require auth.Authorizer) {
	mux.Handle("GET /v1/dashboard/summary", require(auth.PermissionDashboardRead, http.HandlerFunc(h.summary)))
	mux.Handle("GET /v1/dashboard/analytics", require(auth.PermissionDashboardRead, http.HandlerFunc(h.analytics)))
	mux.Handle("GET /v1/dashboard/timeseries", require(auth.PermissionDashboardRead, http.HandlerFunc(h.timeseries)))
	mux.Handle("GET /v1/dashboard/live", require(auth.PermissionDashboardRead, http.HandlerFunc(h.live)))
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.Summary(contextFromRequest(r)))
}

func (h *Handler) analytics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.Analytics(contextFromRequest(r)))
}

func (h *Handler) timeseries(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"points": h.service.TimeSeries(contextFromRequest(r))})
}

func (h *Handler) live(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusOK, h.service.Summary(contextFromRequest(r)))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	context := contextFromRequest(r)
	for index := 0; index < 3; index++ {
		payload, _ := json.Marshal(h.service.Summary(context))
		_, _ = fmt.Fprintf(w, "event: dashboard.summary\ndata: %s\n\n", payload)
		flusher.Flush()
		select {
		case <-r.Context().Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func contextFromRequest(r *http.Request) DashboardContext {
	query := r.URL.Query()
	now := time.Now().UTC()
	from, _ := time.Parse(time.RFC3339, query.Get("from"))
	to, _ := time.Parse(time.RFC3339, query.Get("to"))
	if to.IsZero() {
		to = now
	}
	if from.IsZero() {
		switch query.Get("range") {
		case "7d":
			from = to.Add(-7 * 24 * time.Hour)
		case "30d":
			from = to.Add(-30 * 24 * time.Hour)
		default:
			from = to.Add(-24 * time.Hour)
		}
	}
	currency := query.Get("currency")
	if currency == "" {
		currency = "USD"
	}
	lens := query.Get("analysis_lens")
	if lens == "" {
		lens = "reliability"
	}
	timezone := query.Get("timezone")
	if timezone == "" {
		timezone = "Africa/Lagos"
	}
	return DashboardContext{
		From:         from,
		To:           to,
		Timezone:     timezone,
		ProviderID:   query.Get("provider_id"),
		Corridor:     query.Get("corridor"),
		PayoutMethod: query.Get("payout_method"),
		Currency:     currency,
		AnalysisLens: lens,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
