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

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/ken/connect-microservice/gen/order/v1/orderv1connect"
	"github.com/ken/connect-microservice/internal/config"
	"github.com/ken/connect-microservice/internal/db"
	"github.com/ken/connect-microservice/services/order/internal/client"
	"github.com/ken/connect-microservice/services/order/internal/handler"
	"github.com/ken/connect-microservice/services/order/internal/repository"
	"github.com/ken/connect-microservice/services/order/internal/usecase"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewOrderRepository(pool)
	productClient := client.NewConnectProductClient(cfg.ProductServiceURL)
	userClient := client.NewConnectUserClient(cfg.UserServiceURL)
	uc := usecase.NewOrderUsecase(repo, productClient, userClient)
	h := handler.NewOrderHandler(uc)

	mux := http.NewServeMux()
	path, svcHandler := orderv1connect.NewOrderServiceHandler(h)
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
		slog.Info("order-service starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down order-service")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
