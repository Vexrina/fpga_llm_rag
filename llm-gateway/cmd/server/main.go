package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"llm-gateway/internal/app"
	"llm-gateway/internal/usecases"
	"llm-gateway/internal/utils"
	llm_gateway "llm-gateway/pkg/llm-gateway"
	rag "llm-gateway/pkg/rag"
)

func main() {
	// Configuration
	grpcPort := getEnv("GRPC_PORT", "8083")
	ragServiceAddr := getEnv("RAG_SERVICE_ADDR", "localhost:8082")
	// floatWeaverServiceAddr := getEnv("FLOAT_WEAVER_SERVICE_ADDR", "localhost:8081")
	ollamaApiAddr := getEnv("OLLAMA_API_ADDR", "http://localhost:11434")
	ollamaModel := getEnv("OLLAMA_MODEL", "phi3")

	// gRPC Clients
	ragConn, err := grpc.NewClient(ragServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to rag service: %v", err)
	}
	defer ragConn.Close()
	ragClientGrpc := rag.NewRagServiceClient(ragConn)

	ragClient := utils.NewRagGetter(ragClientGrpc)

	// fwConn, err := grpc.NewClient(floatWeaverServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	log.Fatalf("failed to connect to float-weaver service: %v", err)
	// }
	// defer fwConn.Close()
	// fwClient := floatweaver.NewEmbedServiceClient(fwConn)

	// Ollama Client
	ollamaClient := usecases.NewOllamaClient(ollamaApiAddr)

	// Usecase
	askUsecase := usecases.NewAskUsecase(ragClient, ollamaClient, ollamaModel)

	// gRPC Server
	service := app.NewService(askUsecase)
	server := grpc.NewServer()
	llm_gateway.RegisterGatewayServiceServer(server, service)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC server...")
	server.GracefulStop()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
