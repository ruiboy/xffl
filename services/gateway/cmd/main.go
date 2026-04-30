package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

// logLevelFromEnv returns slog.LevelDebug if LOG_LEVEL=debug, otherwise LevelInfo.
func logLevelFromEnv() slog.Level {
	if os.Getenv("LOG_LEVEL") == "debug" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevelFromEnv(),
	})).With(slog.String("service", "gateway")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	searchURL := os.Getenv("SEARCH_SERVICE_URL")
	if searchURL == "" {
		searchURL = "http://localhost:8082"
	}

	routerURL := os.Getenv("ROUTER_URL")
	if routerURL == "" {
		routerURL = "http://localhost:4000"
	}

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}

	ctx := context.Background()
	searchTarget, err := url.Parse(searchURL)
	if err != nil {
		slog.ErrorContext(ctx, "invalid SEARCH_SERVICE_URL", slog.Any("error", err))
		os.Exit(1)
	}
	searchProxy := httputil.NewSingleHostReverseProxy(searchTarget)

	routerTarget, err := url.Parse(routerURL)
	if err != nil {
		slog.ErrorContext(ctx, "invalid ROUTER_URL", slog.Any("error", err))
		os.Exit(1)
	}
	routerProxy := httputil.NewSingleHostReverseProxy(routerTarget)

	cors := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next(w, r)
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Federated GraphQL endpoint — proxied to Apollo Router.
	mux.HandleFunc("/query", cors(routerProxy.ServeHTTP))
	mux.HandleFunc("/search/query", cors(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/query"
		searchProxy.ServeHTTP(w, r)
	}))

	slog.InfoContext(ctx, "gateway starting", slog.String("port", port), slog.String("router_url", routerURL), slog.String("search_url", searchURL))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.ErrorContext(ctx, "gateway failed", slog.Any("error", err))
		os.Exit(1)
	}
}
