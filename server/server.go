package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joshjms/castletown/config"
	"github.com/joshjms/castletown/sandbox"
	"github.com/joshjms/castletown/server/handler/exec"
)

type Server struct {
	srv *http.Server
	m   *sandbox.Manager
}

func NewServer() (*Server, error) {
	m, err := sandbox.NewManager()
	if err != nil {
		return nil, err
	}

	return &Server{
		m: m,
	}, nil
}

func (s *Server) Init() {
	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: http.DefaultServeMux,
	}
}

func (s *Server) Start() {
	execHandler := exec.NewExecHandler(s.m)

	http.HandleFunc("/exec", execHandler.Handler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		fmt.Printf("Starting server at port %s\n", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down server: %v\n", err)
	}

	fmt.Println("Server gracefully stopped")
}
