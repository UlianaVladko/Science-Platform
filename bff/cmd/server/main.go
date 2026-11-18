// Command server runs the BFF in front of the consolidated literature API.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"scienceplatform/bff/internal/api"
	"scienceplatform/bff/internal/corpus"
	"scienceplatform/bff/internal/enrich"
	"scienceplatform/bff/internal/search"
	"scienceplatform/bff/internal/synthetic"
	"scienceplatform/bff/internal/upstream"
)

const version = "1.0.0"

func main() {
	port := env("PORT", "8080")
	upstreamURL := env("UPSTREAM_URL", "http://127.0.0.1:8000")
	nlpReady := envBool("SYNTHETIC_NLP_READY", false)

	dataset := corpus.Generate()
	engine := search.NewMockEngine(dataset)

	server := &api.Server{
		Upstream:    upstream.New(upstreamURL),
		Corpus:      dataset,
		Engine:      engine,
		Widgets:     synthetic.New(dataset, nlpReady),
		State:       enrich.NewStore(dataset),
		Version:     version,
		UpstreamURL: upstreamURL,
	}

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("BFF listening on :%s (upstream %s, corpus %d publications, engine %s)",
			port, upstreamURL, len(dataset.Publications), engine.Name())
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return parsed
}
