package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-nanobot-litellm-lab/internal/api"
	"go-nanobot-litellm-lab/internal/config"
	"go-nanobot-litellm-lab/internal/litellm"
	"go-nanobot-litellm-lab/internal/router"
	"go-nanobot-litellm-lab/internal/tasks"
)

func main() {
	cfg := config.Load()
	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)
	reviewer, err := litellm.NewClient(litellm.Config{
		BaseURL: cfg.LiteLLMBaseURL,
		APIKey:  cfg.LiteLLMAPIKey,
		Model:   cfg.LiteLLMModel,
		Timeout: cfg.LiteLLMTimeout,
	})
	if err != nil {
		logger.Fatalf("litellm client config failed: %v", err)
	}
	policyRouter, err := router.NewFromFiles(cfg.ModelsConfigPath, cfg.PoliciesConfigPath)
	if err != nil {
		logger.Fatalf("policy router config failed: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.NewHandler(api.Options{Store: tasks.NewStore(), Reviewer: reviewer, PolicyRouter: policyRouter, RequestTimeout: cfg.LiteLLMTimeout, Logger: logger}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Printf("server listening addr=%s", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		logger.Fatalf("server failed: %v", err)
	case sig := <-sigCh:
		logger.Printf("shutdown signal=%s", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("server shutdown failed: %v", err)
	}
}
