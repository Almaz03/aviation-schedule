package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// Создаем контекст с таймаутом (берем из переменной окружения или используем значение по умолчанию)
	dur := getDurationFromEnv("DURATION", 300) // 60 секунд по умолчанию
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dur)*time.Second)
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	registerMetrics()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: ":9100", Handler: mux}

	go func() {
		log.Println("Starting server on :9100")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe failed: %v", err)
		}
	}()

	go startLoad(ctx)

	select {
	case <-stop:
		log.Println("Received shutdown signal")
	case <-ctx.Done():
		log.Println("Timeout reached, shutting down")
	}

	// Гибкое завершение сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server Shutdown failed: %v", err)
	} else {
		log.Println("Server gracefully stopped")
	}
}

func getDurationFromEnv(envVar string, defaultValue int) int {
	if d := os.Getenv(envVar); d != "" {
		if val, err := strconv.Atoi(d); err == nil {
			return val
		}
	}
	return defaultValue
}
