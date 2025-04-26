package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/yourorg/fabric-gateway/proto"
	"google.golang.org/grpc"
)

func main() {
	grpcPort := flag.String("grpc-port", ":50051", "gRPC server port")
	httpPort := flag.String("http-port", ":8080", "REST gateway port")
	flag.Parse()

	// 1) Start gRPC server
	lis, err := net.Listen("tcp", *grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterFabricServiceServer(grpcServer, &FabricHandler{})
	go func() {
		log.Printf("gRPC server listening on %s", *grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	// 2) Start HTTP REST Gateway
	ctx := context.Background()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()} // TLS config as needed
	err = pb.RegisterFabricServiceHandlerFromEndpoint(ctx, mux, *grpcPort, opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}
	log.Printf("REST gateway listening on %s", *httpPort)
	log.Fatal(http.ListenAndServe(*httpPort, mux))
}
