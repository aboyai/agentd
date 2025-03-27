// cmd/server/main.go
package main

import (
	"agentd/api"
	"agentd/proto"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	// Start Visualizer before blocking gRPC Serve()
	go StartVisualizerServer()

	// Start REST API Gateway
	go StartRestGateway()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterInstructionServiceServer(grpcServer, &api.AgentDHandler{})

	fmt.Println("🚀 AgentD is running on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
