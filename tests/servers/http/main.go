package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	port       = getEnv("MOCK_HTTP_PORT", "18080")
	retryCount atomic.Int32
	slowDelay  = getEnvDuration("MOCK_SLOW_DELAY", 5*time.Second)
)

func main() {
	mux := http.NewServeMux()
	registerRoutes(mux)

	addr := ":" + port
	log.Printf("Mock HTTP server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func registerRoutes(mux *http.ServeMux) {
	// Success endpoints for all methods
	mux.HandleFunc("/get", handlerGet)
	mux.HandleFunc("/post", handlerPost)
	mux.HandleFunc("/put", handlerPut)
	mux.HandleFunc("/patch", handlerPatch)
	mux.HandleFunc("/delete", handlerDelete)
	mux.HandleFunc("/head", handlerHead)
	mux.HandleFunc("/options", handlerOptions)

	// Error scenario endpoints
	mux.HandleFunc("/timeout", handlerTimeout)
	mux.HandleFunc("/retry", handlerRetry)
	mux.HandleFunc("/fail", handlerFail)
	mux.HandleFunc("/error", handlerError)
	mux.HandleFunc("/slow", handlerSlow)

	// Utility endpoints
	mux.HandleFunc("/echo", handlerEcho)
	mux.HandleFunc("/health", handlerHealth)
	mux.HandleFunc("/status/", handlerStatus)
}

// --- Success Handlers ---

func handlerGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"method":  "GET",
		"message": "success",
		"query":   r.URL.Query(),
	})
}

func handlerPost(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"method":  "POST",
		"message": "created",
		"body":    string(body),
		"headers": headerMap(r.Header),
	})
}

func handlerPut(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"method":  "PUT",
		"message": "updated",
		"body":    string(body),
	})
}

func handlerPatch(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"method":  "PATCH",
		"message": "patched",
		"body":    string(body),
	})
}

func handlerDelete(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"method":  "DELETE",
		"message": "deleted",
	})
}

func handlerHead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Custom-Header", "head-response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func handlerOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
	w.WriteHeader(http.StatusOK)
}

// --- Error Scenario Handlers ---

func handlerTimeout(w http.ResponseWriter, r *http.Request) {
	select {
	case <-time.After(slowDelay):
		writeJSON(w, http.StatusOK, map[string]string{"message": "delayed"})
	case <-r.Context().Done():
		return
	}
}

func handlerRetry(w http.ResponseWriter, r *http.Request) {
	count := retryCount.Add(1)
	if count <= 2 {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error":   "temporary failure",
			"attempt": count,
		})
		return
	}
	retryCount.Store(0)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "success after retry",
		"attempts": count,
	})
}

func handlerFail(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "permanent failure",
	})
}

func handlerError(w http.ResponseWriter, r *http.Request) {
	codeStr := r.URL.Query().Get("code")
	code := http.StatusInternalServerError
	if codeStr != "" {
		if c, err := strconv.Atoi(codeStr); err == nil {
			code = c
		}
	}
	writeJSON(w, code, map[string]interface{}{
		"error": fmt.Sprintf("error with code %d", code),
	})
}

func handlerSlow(w http.ResponseWriter, r *http.Request) {
	delayStr := r.URL.Query().Get("delay")
	delay := slowDelay
	if delayStr != "" {
		if d, err := time.ParseDuration(delayStr); err == nil {
			delay = d
		}
	}
	time.Sleep(delay)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "slow response",
		"delay":   delay.String(),
	})
}

// --- Utility Handlers ---

func handlerEcho(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"method":  r.Method,
		"url":     r.URL.String(),
		"headers": headerMap(r.Header),
		"body":    string(body),
	})
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func handlerStatus(w http.ResponseWriter, r *http.Request) {
	codeStr := r.URL.Path[len("/status/"):]
	code := 200
	if codeStr != "" {
		if c, err := strconv.Atoi(codeStr); err == nil {
			code = c
		}
	}
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf(`{"status":%d}`, code)))
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func headerMap(h http.Header) map[string]string {
	m := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
