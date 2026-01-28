package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/esivanov203/antibruteforce/internal/httpapi"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("server starting")

	if err := godotenv.Load(); err != nil {
		log.Printf(".env file not found or failed to load: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	app, err := appFactory(
		os.Getenv("LIMIT_DURATION"),
		os.Getenv("LIMIT_LOGIN"),
		os.Getenv("LIMIT_PASSWORD"),
		os.Getenv("LIMIT_IP"),
	)
	if err != nil {
		log.Println("app initial:", err)
	}

	s := httpapi.New(os.Getenv("INNER_HTTP_PORT"), app)

	stopCh := make(chan struct{}, 1)
	go s.Start(stopCh)

	// ждем системный сигнал отмены или ошибку запуска сервера
	select {
	case <-ctx.Done():
	case <-stopCh:
		cancel()
	}

	// graceful shutdown
	shdCtx, shdCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shdCancel()
	s.Stop(shdCtx)

	log.Println("server stopped")
}
