package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"llm-gateway/internal/app"
	"llm-gateway/internal/usecases"
	"llm-gateway/internal/utils"
	llm_gateway "llm-gateway/pkg/llm-gateway"
	rag "llm-gateway/pkg/rag"
)

func main() {
	// Configuration
	grpcPort := getEnv("GRPC_PORT", "8083")
	ragServiceAddr := getEnv("RAG_SERVICE_ADDR", "localhost:50051")
	// floatWeaverServiceAddr := getEnv("FLOAT_WEAVER_SERVICE_ADDR", "localhost:8081")
	ollamaApiAddr := getEnv("OLLAMA_API_ADDR", "http://localhost:11434")
	ollamaModel := getEnv("OLLAMA_MODEL", "phi3:latest")

	fmt.Println(ragServiceAddr)
	// gRPC Clients
	ragConn := getConn("localhost", "50051")
	defer ragConn.Close()

	// Пробуем сделать первый вызов, чтобы проверить соединение:
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := rag.NewRagServiceClient(ragConn).SearchDocuments(ctx, &rag.SearchRequest{Query: "hello", Limit: 10, SimilarityThreshold: 0.5})
	if err != nil {
		log.Fatalf("Ошибка при вызове rag: %v", err)
	}
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
	reflection.Register(server)

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

func getConn(url, port string) *grpc.ClientConn {
	log.Println("URL:", url)
	log.Println("PORT:", port)
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", url, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Не создали коннекшн с floatweaver: %v", err)
	}
	return conn
}
