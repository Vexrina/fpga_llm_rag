package main

import (
	"float-weaver/internal/usecases"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"float-weaver/internal/app"
	pb "float-weaver/pkg/floatweaver"
)

func main() {
	port := getEnv("PORT", "8081")
	ollamaUrl := getEnv("OLLAMA_URL", "http://localhost:11434")
	model := getEnv("OLLAMA_EMBEDDING_MODEL", "bge-m3")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Не удалось создать listener: %v", err)
	}

	s := grpc.NewServer()

	embedUc := usecases.NewEmbedUsecase(ollamaUrl, model)
	log.Printf("Using Ollama at %s with model: %s", ollamaUrl, model)

	ragServer := app.NewFloatWeaver(embedUc)
	pb.RegisterEmbedServiceServer(s, ragServer)

	reflection.Register(s)

	log.Printf("gRPC сервер запущен на порту %s", port)

	if err = s.Serve(lis); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
