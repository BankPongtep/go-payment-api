package payment

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/BankPongtep/go-payment-api/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	db *db.DB
}

func NewHandler(database *db.DB) *Handler {
	return &Handler{db: database}
}

// CheckoutRequest is the payload for creating a payment
type CheckoutRequest struct {
	Amount         int64  `json:"amount"`           // in satangs (THB)
	Currency       string `json:"currency"`         // "THB"
	Token          string `json:"token"`            // Omise card token
	Description    string `json:"description"`
	IdempotencyKey string `json:"idempotency_key"`
}

// PaymentResponse is returned to the client
type PaymentResponse struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Amount      int64     `json:"amount"`
	Currency    string    `json:"currency"`
	Paid        bool      `json:"paid"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Checkout creates a new payment charge
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		jsonError(w, "amount must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.Token == "" {
		jsonError(w, "token is required", http.StatusBadRequest)
		return
	}

	// TODO: Call Omise API to create charge
	// charge, err := omiseClient.CreateCharge(req.Amount, req.Token, req.Description)
	log.Info().
		Int64("amount", req.Amount).
		Str("currency", req.Currency).
		Str("idempotency_key", req.IdempotencyKey).
		Msg("checkout initiated")

	// Mock response (replace with real Omise call)
	resp := PaymentResponse{
		ID:          "chrg_test_" + req.IdempotencyKey,
		Status:      "successful",
		Amount:      req.Amount,
		Currency:    req.Currency,
		Paid:        true,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetPayment retrieves a payment by ID
func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// TODO: Query from database
	resp := PaymentResponse{
		ID:        id,
		Status:    "successful",
		Amount:    150000,
		Currency:  "THB",
		Paid:      true,
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Refund refunds a payment
func (h *Handler) Refund(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	log.Info().Str("payment_id", id).Msg("refund initiated")

	// TODO: Call Omise refund API
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         "rfnd_test_xxx",
		"charge_id":  id,
		"status":     "pending",
		"created_at": time.Now(),
	})
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
