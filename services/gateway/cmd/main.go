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
	})))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	aflURL := os.Getenv("AFL_SERVICE_URL")
	if aflURL == "" {
		aflURL = "http://localhost:8080"
	}

	fflURL := os.Getenv("FFL_SERVICE_URL")
	if fflURL == "" {
		fflURL = "http://localhost:8081"
	}

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}

	ctx := context.Background()
	aflTarget, err := url.Parse(aflURL)
	if err != nil {
		slog.ErrorContext(ctx, "invalid AFL_SERVICE_URL", slog.Any("error", err))
		os.Exit(1)
	}
	aflProxy := httputil.NewSingleHostReverseProxy(aflTarget)

	fflTarget, err := url.Parse(fflURL)
	if err != nil {
		slog.ErrorContext(ctx, "invalid FFL_SERVICE_URL", slog.Any("error", err))
		os.Exit(1)
	}
	fflProxy := httputil.NewSingleHostReverseProxy(fflTarget)

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

	mux.HandleFunc("/query", cors(aflProxy.ServeHTTP))
	mux.HandleFunc("/afl/query", cors(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/query"
		aflProxy.ServeHTTP(w, r)
	}))
	mux.HandleFunc("/ffl/query", cors(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/query"
		fflProxy.ServeHTTP(w, r)
	}))

	slog.InfoContext(ctx, "gateway starting", slog.String("port", port), slog.String("afl_url", aflURL), slog.String("ffl_url", fflURL))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.ErrorContext(ctx, "gateway failed", slog.Any("error", err))
		os.Exit(1)
	}
}
