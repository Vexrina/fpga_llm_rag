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
	port := getEnv("PORT", "8080")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Не удалось создать listener: %v", err)
	}

	// Создаем gRPC сервер
	s := grpc.NewServer()

	// Создаем юзкейсы
	embedUc := usecases.NewEmbedUsecase()

	// Регистрируем наш сервис
	ragServer := app.NewFloatWeaver(embedUc)
	pb.RegisterEmbedServiceServer(s, ragServer)

	// Включаем reflection для отладки
	reflection.Register(s)

	log.Printf("gRPC сервер запущен на порту %s", port)

	// Запускаем сервер
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
