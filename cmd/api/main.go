package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BankPongtep/go-payment-api/internal/auth"
	"github.com/BankPongtep/go-payment-api/internal/db"
	"github.com/BankPongtep/go-payment-api/internal/middleware"
	"github.com/BankPongtep/go-payment-api/internal/payment"
	"github.com/BankPongtep/go-payment-api/internal/webhook"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load .env
	_ = godotenv.Load()

	// Logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	// DB
	database, err := db.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}
	defer database.Close()

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.RateLimiter)

	// Health & metrics
	r.Get("/health", healthHandler)
	r.Handle("/metrics", promhttp.Handler())

	// Auth routes
	authHandler := auth.NewHandler(database)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	// Payment routes (protected)
	paymentHandler := payment.NewHandler(database)
	r.Route("/api/v1/payments", func(r chi.Router) {
		r.Use(middleware.JWTAuth)
		r.Post("/checkout", paymentHandler.Checkout)
		r.Get("/{id}", paymentHandler.GetPayment)
		r.Post("/{id}/refund", paymentHandler.Refund)
	})

	// Webhook routes
	webhookHandler := webhook.NewHandler(database)
	r.Route("/api/v1/webhooks", func(r chi.Router) {
		r.Post("/omise", webhookHandler.HandleOmise)
	})

	// Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info().Msgf("🚀 Server running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("forced shutdown")
	}
	log.Info().Msg("Server stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"go-payment-api","version":"1.0.0"}`))
}
