package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/BankPongtep/go-payment-api/internal/db"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	db *db.DB
}

func NewHandler(database *db.DB) *Handler {
	return &Handler{db: database}
}

type OmiseEvent struct {
	Key  string          `json:"key"`
	Data json.RawMessage `json:"data"`
}

// HandleOmise processes Omise webhook events
func (h *Handler) HandleOmise(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"cannot read body"}`, http.StatusBadRequest)
		return
	}

	// Verify HMAC signature
	if !verifySignature(body, r.Header.Get("OmiseKey")) {
		log.Warn().Msg("webhook: invalid signature")
		http.Error(w, `{"error":"invalid signature"}`, http.StatusUnauthorized)
		return
	}

	var event OmiseEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, `{"error":"invalid payload"}`, http.StatusBadRequest)
		return
	}

	log.Info().Str("event_key", event.Key).Msg("webhook received")

	switch event.Key {
	case "charge.complete":
		h.handleChargeComplete(event.Data)
	case "charge.create":
		h.handleChargeCreate(event.Data)
	default:
		log.Info().Str("event_key", event.Key).Msg("unhandled event")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"received":true}`))
}

func (h *Handler) handleChargeComplete(data json.RawMessage) {
	// TODO: Update payment status in DB
	log.Info().RawJSON("data", data).Msg("charge completed")
}

func (h *Handler) handleChargeCreate(data json.RawMessage) {
	log.Info().RawJSON("data", data).Msg("charge created")
}

func verifySignature(payload []byte, signature string) bool {
	secret := os.Getenv("OMISE_WEBHOOK_SECRET")
	if secret == "" {
		return true // skip in dev
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
