package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joshjms/castletown/config"
	pb "github.com/joshjms/castletown/proto"
	"github.com/joshjms/castletown/server/handler/done"
	"github.com/joshjms/castletown/server/handler/exec"
	"google.golang.org/grpc"
)

type Server struct {
	httpSrv *http.Server
	grpcSrv *grpc.Server
}

func NewServer() (*Server, error) {
	grpcSrv := grpc.NewServer()

	pb.RegisterExecServiceServer(grpcSrv, exec.NewExecServer())
	pb.RegisterDoneServiceServer(grpcSrv, done.NewDoneServer())

	return &Server{
		httpSrv: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.Port),
			Handler: http.DefaultServeMux,
		},
		grpcSrv: grpcSrv,
	}, nil
}

func (s *Server) Start() {
	http.HandleFunc("/exec", exec.Handler)
	http.HandleFunc("/done", done.Handler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		fmt.Printf("Starting HTTP server at port %s\n", s.httpSrv.Addr)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting HTTP server: %v\n", err)
		}
	}()

	grpcPort := config.Port + 1
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			fmt.Printf("Failed to listen for gRPC: %v\n", err)
			return
		}
		fmt.Printf("Starting gRPC server at port %d\n", grpcPort)
		if err := s.grpcSrv.Serve(lis); err != nil {
			fmt.Printf("Error starting gRPC server: %v\n", err)
		}
	}()

	<-stop

	fmt.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpSrv.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down HTTP server: %v\n", err)
	}

	s.grpcSrv.GracefulStop()

	fmt.Println("Servers gracefully stopped")
}
