package intake

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/CloudSpaceLab/imsi-rails/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/transactions", h.submitTransaction)
	mux.HandleFunc("GET /v1/transactions/{transaction_id}", h.getTransaction)
}

func (h *Handler) RegisterProtected(mux *http.ServeMux, require auth.Authorizer) {
	mux.Handle("POST /v1/transactions", require(auth.PermissionTransactionsTrace, http.HandlerFunc(h.submitTransaction)))
	mux.Handle("GET /v1/transactions/{transaction_id}", require(auth.PermissionTransactionsRead, http.HandlerFunc(h.getTransaction)))
}

func (h *Handler) submitTransaction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request SubmitTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	response, err := h.service.Submit(request)
	if err != nil {
		writeProblem(w, statusForValidationError(err), "invalid_transaction", err.Error())
		return
	}

	status := http.StatusCreated
	if response.IDempotentReplay {
		status = http.StatusOK
	}
	writeJSON(w, status, response)
}

func (h *Handler) getTransaction(w http.ResponseWriter, r *http.Request) {
	transactionID := strings.TrimSpace(r.PathValue("transaction_id"))
	if transactionID == "" {
		writeProblem(w, http.StatusBadRequest, "missing_transaction_id", "transaction_id is required.")
		return
	}

	record, ok := h.service.transactions.GetByTransactionID(transactionID)
	if !ok {
		writeProblem(w, http.StatusNotFound, "transaction_not_found", "Transaction was not found.")
		return
	}

	writeJSON(w, http.StatusOK, record)
}

func statusForValidationError(err error) int {
	switch {
	case errors.Is(err, ErrMissingIDempotencyKey),
		errors.Is(err, ErrMissingBankID),
		errors.Is(err, ErrMissingCurrency),
		errors.Is(err, ErrInvalidAmount),
		errors.Is(err, ErrMissingPayoutMethod),
		errors.Is(err, ErrMissingCountries):
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
