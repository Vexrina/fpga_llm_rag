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
	llm_gateway "rag/pkg/llm-gateway"
	pb "rag/pkg/rag"
)

const (
	defaultPDFTimeout        = 5 * time.Minute
	defaultLinkScrapeTimeout = 3 * time.Minute
)

type LLMGatewayClient struct {
	conn *grpc.ClientConn
}

func NewLLMGatewayClient(addr string) (*LLMGatewayClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to llm-gateway: %w", err)
	}
	return &LLMGatewayClient{conn: conn}, nil
}

func (c *LLMGatewayClient) NotifyBasePromptUpdate(newPrompt string) {
	if c.conn == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client := llm_gateway.NewGatewayServiceClient(c.conn)
		_, err := client.UpdateBasePrompt(ctx, &llm_gateway.UpdateBasePromptRequest{Prompt: newPrompt})
		if err != nil {
			log.Printf("Failed to push basePrompt to LLM Gateway: %v", err)
		} else {
			log.Printf("Successfully pushed basePrompt update to LLM Gateway")
		}
	}()
}

func (c *LLMGatewayClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

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
		llmGatewayAddr = getEnv("LLM_GATEWAY_ADDR", "llm-gateway:8083")
	)

	var ( // conns
		fwHost = getEnv("FLOATWEAVER_HOST", "float-weaver")
		fwPort = getEnv("FLOATWEAVER_PORT", "8081")
		fwConn = getConn(fwHost, fwPort)
	)

	defer fwConn.Close()

	var ( // grpc clienst
		fw = floatweaver.NewEmbedServiceClient(fwConn)
	)

	// only for database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db := repository.NewVecDb(ctx, connStr)

	// LLM Gateway client for push notifications
	llmClient, err := NewLLMGatewayClient(llmGatewayAddr)
	if err != nil {
		log.Printf("Warning: Could not connect to LLM Gateway for push notifications: %v", err)
	}
	if llmClient != nil {
		defer llmClient.Close()
	}

	var (
		pdfProcessor = usecases.NewPDFProcessor(
			"python3",
			"/app/python_scripts/kug_scrap.py",
			[]string{},
			defaultPDFTimeout,
		)
		linkScraperProcessor = usecases.NewLinkScraperProcessor(
			"python3",
			"/app/python_scripts/link_scraper_cached.py",
			"/app/.web_cache/link_scraper",
			defaultLinkScrapeTimeout,
		)
		addDocumentUsecase     = usecases.NewAddDocumentUsecase(db, fw, pdfProcessor, linkScraperProcessor)
		previewDocumentUsecase = usecases.NewPreviewDocumentUsecase(pdfProcessor, linkScraperProcessor)
		commitDocumentUsecase  = usecases.NewCommitDocumentUsecase(db, fw)
		settingsUsecase        = usecases.NewSettingsUsecase(db, llmClient)
		searchDocumentUsecase  = usecases.NewSearchDocumentsUsecase(db, fw, settingsUsecase, db)
		documentHistoryUsecase = usecases.NewDocumentHistoryUsecase(db)
		queryLogsUsecase       = usecases.NewQueryLogsUsecase(db)
		discoverLinksUsecase   = usecases.NewDiscoverLinksUsecase(linkScraperProcessor)
		scrapeUrlsUsecase      = usecases.NewScrapeUrlsUsecase(linkScraperProcessor)
	)

	// Создаем TCP listener на порту 50051
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Не удалось создать listener: %v", err)
	}

	// Создаем gRPC сервер с увеличенным лимитом для больших PDF (20MB * 1.33 ~= 27MB base64)
	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(27*1024*1024),
		grpc.MaxSendMsgSize(27*1024*1024),
	)

	// Регистрируем наш сервис
	ragServer := app.NewRagServer(
		db,
		addDocumentUsecase,
		previewDocumentUsecase,
		commitDocumentUsecase,
		searchDocumentUsecase,
		settingsUsecase,
		documentHistoryUsecase,
		queryLogsUsecase,
		discoverLinksUsecase,
		scrapeUrlsUsecase,
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
