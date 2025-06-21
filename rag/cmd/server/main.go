package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"rag/internal"
	pb "rag/pkg/rag"
)

// some check
func main() {
	// Создаем TCP listener на порту 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Не удалось создать listener: %v", err)
	}

	// Создаем gRPC сервер
	s := grpc.NewServer()

	// Регистрируем наш сервис
	ragServer := internal.NewRagServer()
	pb.RegisterRagServiceServer(s, ragServer)

	// Включаем reflection для отладки
	reflection.Register(s)

	log.Printf("gRPC сервер запущен на порту 50051")

	// Запускаем сервер
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
