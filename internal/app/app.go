package app

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/437d5/subpub/internal/config"
	"github.com/437d5/subpub/internal/server"
	"github.com/437d5/subpub/pkg/pb"
	"github.com/437d5/subpub/pkg/subpub"
	"google.golang.org/grpc"
)

type App struct {
	cfg          config.Config
	pubsub       subpub.SubPub
	grpcServer   *grpc.Server
	shutdownChan chan struct{}
}

func New(cfg config.Config, bus subpub.SubPub) *App {
	return &App{
		cfg:          cfg,
		pubsub:       bus,
		shutdownChan: make(chan struct{}),
	}
}

func (a *App) Run() error {
	a.grpcServer = grpc.NewServer()
	pb.RegisterPubSubServer(a.grpcServer, server.Server{
		Subpub:       a.pubsub,
		ShutdownChan: a.shutdownChan,
	})

	if a.cfg.GRPCPort == "" {
		a.cfg.GRPCPort = "5000"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", a.cfg.GRPCPort))
	if err != nil {
		return err
	}

	log.Printf("Starting gRPC server on port %s", a.cfg.GRPCPort)
	return a.grpcServer.Serve(lis)
}

func (a *App) Stop(ctx context.Context) error {
	log.Println("Starting graceful shutdown...")
	close(a.shutdownChan)

	if err := a.pubsub.Close(ctx); err != nil {
		log.Printf("SubPub close failed: %v", err)
		return err
	}

	stopped := make(chan struct{})
	defer close(stopped)

	go func() {
		a.grpcServer.GracefulStop()
		stopped <- struct{}{}
	}()

	select {
	case <-stopped:
		log.Println("Server stopped gracefully")
		return nil
	case <-ctx.Done():
		log.Println("Shutdown due to timeout")
		a.grpcServer.Stop()
		return ctx.Err()
	}
}
