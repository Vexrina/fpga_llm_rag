package usecases

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"rag/internal/processors"
	"rag/internal/repository"
	"rag/internal/utils"
	"rag/pkg/floatweaver/floatweaver"
)

type PDFProcessor interface {
	ExtractTextFromPDF(pdfData []byte) (string, error)
}

type URLProcessor interface {
	ExtractTextFromURL(url string, maxDepth int32) (string, error)
	DiscoverLinks(url string, maxDepth int32) ([]string, error)
}

func NewPDFProcessor(pythonPath, scriptPath string, pythonArgs []string, timeout time.Duration) PDFProcessor {
	return processors.NewKugScrapProcessor(pythonPath, scriptPath, pythonArgs, timeout)
}

func NewLinkScraperProcessor(pythonPath, scriptPath, cacheDir string, timeout time.Duration) URLProcessor {
	return processors.NewLinkScraperProcessor(pythonPath, scriptPath, cacheDir, timeout)
}

type AddDocumentRepository interface {
	InsertItemWithTx(ctx context.Context, tx pgx.Tx, item repository.Item) error
	WithTransactional(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type AddDocumentUsecase struct {
	addDocRepository  AddDocumentRepository
	floatWeaverClient floatweaver.EmbedServiceClient
	pdfProcessor      PDFProcessor
	urlProcessor      URLProcessor
}

func NewAddDocumentUsecase(
	repository AddDocumentRepository,
	floatWeaverClient floatweaver.EmbedServiceClient,
	pdfProcessor PDFProcessor,
	urlProcessor URLProcessor,
) *AddDocumentUsecase {
	return &AddDocumentUsecase{
		addDocRepository:  repository,
		floatWeaverClient: floatWeaverClient,
		pdfProcessor:      pdfProcessor,
		urlProcessor:      urlProcessor,
	}

}

func (u *AddDocumentUsecase) AddDocument(ctx context.Context, domain *utils.AddDocumentDomain) error {
	content := domain.Content

	switch domain.SourceType {
	case utils.DocumentSourceTypeURL:
		if u.urlProcessor == nil {
			return fmt.Errorf("url processor not configured")
		}
		if domain.SourceURL == "" {
			return fmt.Errorf("source_url is required for URL source type")
		}
		text, err := u.urlProcessor.ExtractTextFromURL(domain.SourceURL, domain.URLMaxDepth)
		if err != nil {
			return fmt.Errorf("urlProcessor.ExtractTextFromURL got error: %w", err)
		}
		content = text

	case utils.DocumentSourceTypePDF:
		if u.pdfProcessor == nil {
			return fmt.Errorf("pdf processor not configured")
		}
		pdfData, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return fmt.Errorf("failed to decode PDF from base64: %w", err)
		}
		if len(pdfData) > 20*1024*1024 {
			return fmt.Errorf("pdf size exceeds 20MB limit")
		}
		text, err := u.pdfProcessor.ExtractTextFromPDF(pdfData)
		if err != nil {
			return fmt.Errorf("pdfProcessor.ExtractTextFromPDF got error: %w", err)
		}
		content = text
	}

	embed, err := u.floatWeaverClient.Embed(ctx, &floatweaver.EmbedRequest{Text: content})
	if err != nil {
		return fmt.Errorf("floatWeaverClient.Embed got error: %w", err)
	}
	return u.addDocRepository.WithTransactional(ctx, func(tx pgx.Tx) error {
		item := repository.Item{
			Title:     domain.Title,
			Embedding: embed.Embeddings[0].Values,
			Text:      content,
			Metadata:  domain.Metadata,
		}
		return u.addDocRepository.InsertItemWithTx(ctx, tx, item)
	})
}
