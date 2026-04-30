package main

import (
	"context"
	"errors"
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

	"github.com/ken/connect-microservice/gen/user/v1/userv1connect"
	"github.com/ken/connect-microservice/internal/auth"
	"github.com/ken/connect-microservice/internal/config"
	"github.com/ken/connect-microservice/internal/db"
	"github.com/ken/connect-microservice/internal/middleware"
	"github.com/ken/connect-microservice/services/user/internal/handler"
	"github.com/ken/connect-microservice/services/user/internal/repository"
	"github.com/ken/connect-microservice/services/user/internal/usecase"
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

	repo := repository.NewUserRepository(pool)
	uc := usecase.NewUserUsecase(repo)
	h := handler.NewUserHandler(uc, tokenGen)

	// Login と CreateUser は認証不要（public エンドポイント）
	publicProcedures := []string{
		userv1connect.UserServiceLoginProcedure,
		userv1connect.UserServiceCreateUserProcedure,
	}
	authInterceptor := middleware.NewAuthInterceptor(tokenGen, publicProcedures)

	mux := http.NewServeMux()
	path, svcHandler := userv1connect.NewUserServiceHandler(h, connect.WithInterceptors(authInterceptor))
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
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("user-service starting", "addr", addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		slog.Error("server error", "error", err)
		os.Exit(1)
	case <-quit:
	}

	slog.Info("shutting down user-service")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
