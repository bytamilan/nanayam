package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/bytamilan/nanayam/services/gateway/proto"
	"google.golang.org/grpc"
)

func main() {
	grpcPort := flag.String("grpc-port", ":50051", "gRPC server port")
	httpPort := flag.String("http-port", ":8080", "REST server port")
	flag.Parse()

	// Load gateway configuration from environment
	cfg := NewGatewayFromEnv()
	log.Printf("Connecting to Fabric peer at %s (MSP: %s)", cfg.PeerEndpoint, cfg.MSP_ID)

	// Connect to Fabric
	gw, conn, err := cfg.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Fabric gateway: %v", err)
	}
	defer conn.Close()
	defer gw.Close()

	log.Println("Connected to Fabric network successfully")

	// Initialize auth store
	authStore := NewAuthStore()
	authStore.SeedAdmin()
	if authStore.IsSignupEnabled() {
		log.Println("User registration is ENABLED")
	} else {
		log.Println("User registration is DISABLED")
	}

	handler := NewFabricHandler(gw, cfg.ChannelName, cfg.ChaincodeName)

	// 1) Start gRPC server
	lis, err := net.Listen("tcp", *grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", *grpcPort, err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterFabricServiceServer(grpcServer, handler)

	go func() {
		log.Printf("gRPC distribution server listening on %s", *grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	// 2) Start HTTP REST server
	rest := &RESTServer{
		handler:   handler,
		authStore: authStore,
		cfg:       cfg,
	}
	mux := http.NewServeMux()
	rest.register(mux)

	httpServer := &http.Server{
		Addr:    *httpPort,
		Handler: mux,
	}

	go func() {
		log.Printf("REST distribution gateway listening on %s", *httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP serve error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down distribution server...")
	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(context.Background())
}
