package main

import (
	"log"
	"net"
	"os"
	"rag/internal/app"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "rag/pkg/rag"
)

// some check
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Создаем TCP listener на порту 50051
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Не удалось создать listener: %v", err)
	}

	// Создаем gRPC сервер
	s := grpc.NewServer()

	// Регистрируем наш сервис
	ragServer := app.NewRagServer()
	pb.RegisterRagServiceServer(s, ragServer)

	// Включаем reflection для отладки
	reflection.Register(s)

	log.Printf("gRPC сервер запущен на порту %s", port)

	// Запускаем сервер
	if err = s.Serve(lis); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
