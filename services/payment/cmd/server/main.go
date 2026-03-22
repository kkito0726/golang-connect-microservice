package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/ken/connect-microservice/gen/payment/v1/paymentv1connect"
	"github.com/ken/connect-microservice/internal/auth"
	"github.com/ken/connect-microservice/internal/config"
	"github.com/ken/connect-microservice/internal/db"
	"github.com/ken/connect-microservice/internal/middleware"
	"github.com/ken/connect-microservice/services/payment/internal/client"
	"github.com/ken/connect-microservice/services/payment/internal/handler"
	"github.com/ken/connect-microservice/services/payment/internal/repository"
	"github.com/ken/connect-microservice/services/payment/internal/usecase"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	tokenGen := auth.NewTokenGenerator(cfg.JWTSecret, cfg.JWTExpiryHours)
	authInterceptor := middleware.NewAuthInterceptor(tokenGen, nil)

	repo := repository.NewPaymentRepository(pool)
	orderClient := client.NewConnectOrderClient(cfg.OrderServiceURL)
	uc := usecase.NewPaymentUsecase(repo, orderClient)
	h := handler.NewPaymentHandler(uc)

	mux := http.NewServeMux()
	path, svcHandler := paymentv1connect.NewPaymentServiceHandler(h, connect.WithInterceptors(authInterceptor))
	mux.Handle(path, svcHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("payment-service starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down payment-service")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
