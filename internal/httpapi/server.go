package httpapi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/esivanov203/antibruteforce/internal/service"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Server struct {
	httpServer *http.Server
	router     *mux.Router
	app        service.App
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func New(port string) *Server {
	router := mux.NewRouter()
	s := &Server{
		httpServer: &http.Server{
			Addr:              ":" + port,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       30 * time.Second,
		},
		router: router,
	}
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	s.router.Use(s.logMiddleware)
	s.setRowerHandlers()

	return s
}

func (s *Server) Start(chanErr chan struct{}) {
	// todo info
	log.Printf("http server starting on: %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// todo err
		log.Printf("http server not started: %v", err)
	}

	// если сервер завершился сигнализируем
	select {
	case chanErr <- struct{}{}:
	default:
	}
}

func (s *Server) Stop(ctx context.Context) {
	// todo info
	log.Printf("http server %s stopping", s.httpServer.Addr)
	if err := s.httpServer.Shutdown(ctx); err != nil {
		// todo err
		fmt.Printf("http server stopped error: %v", err)
	}
}

func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		next.ServeHTTP(rw, r)

		// todo info
		log.Printf("HTTP request - method: %s, ip: %s, path: %s, status: %d, latency: %d, uagent: %s, proto: %s",
			r.Method, r.RemoteAddr, r.URL.Path, rw.status, time.Since(start).Milliseconds(), r.UserAgent(), r.Proto)
	})
}
