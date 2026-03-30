package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
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

	aflTarget, err := url.Parse(aflURL)
	if err != nil {
		log.Fatalf("invalid AFL_SERVICE_URL: %v", err)
	}
	aflProxy := httputil.NewSingleHostReverseProxy(aflTarget)

	fflTarget, err := url.Parse(fflURL)
	if err != nil {
		log.Fatalf("invalid FFL_SERVICE_URL: %v", err)
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
	mux.HandleFunc("/afl/query", cors(aflProxy.ServeHTTP))
	mux.HandleFunc("/ffl/query", cors(fflProxy.ServeHTTP))

	log.Printf("Gateway starting on :%s (AFL→%s, FFL→%s)", port, aflURL, fflURL)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
