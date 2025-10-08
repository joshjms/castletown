package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joshjms/castletown/config"
	"github.com/joshjms/castletown/server/handler/exec"
	"github.com/joshjms/castletown/server/handler/finish"
)

type Server struct {
	srv *http.Server
}

func NewServer() (*Server, error) {
	return &Server{
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.Port),
			Handler: http.DefaultServeMux,
		},
	}, nil
}

func (s *Server) Start() {
	http.HandleFunc("/exec", exec.Handler)
	http.HandleFunc("/finish", finish.Handler)

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
