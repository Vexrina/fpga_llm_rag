package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"rag/internal/repository"
	"rag/internal/usecases"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"rag/internal/app"
	"rag/pkg/floatweaver/floatweaver"
	pb "rag/pkg/rag"
)

// some check
func main() {
	var ( // envs
		port       = getEnv("PORT", "8083")
		dbUser     = getEnv("POSTGRES_USER", "rag_user")
		dbPassword = getEnv("POSTGRES_PASSWORD", "rag_password")
		dbHost     = getEnv("POSTGRES_HOST", "localhost")
		dbPort     = getEnv("POSTGRES_PORT", "5432")
		dbName     = getEnv("POSTGRES_DB", "rag_db")
		connStr    = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPassword, dbHost, dbPort, dbName,
		)
	)

	var ( // conns
		fwConn = getConn("localhost", "8081")
	)

	defer fwConn.Close()

	var ( // grpc clienst
		fw = floatweaver.NewEmbedServiceClient(fwConn)
	)

	// only for database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db := repository.NewVecDb(ctx, connStr)

	var (
		addDocumentUsecase = usecases.NewAddDocumentUsecase(db, fw)
	)

	// Создаем TCP listener на порту 50051
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Не удалось создать listener: %v", err)
	}

	// Создаем gRPC сервер
	s := grpc.NewServer()

	// Регистрируем наш сервис
	ragServer := app.NewRagServer(
		db,
		addDocumentUsecase,
	)
	pb.RegisterRagServiceServer(s, ragServer)

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

func getConn(url, port string) *grpc.ClientConn {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", url, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Не создали коннекшн с floatweaver: %v", err)
	}
	return conn
}
