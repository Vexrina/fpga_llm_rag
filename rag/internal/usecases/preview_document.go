package usecases

import (
	"context"
	"encoding/base64"
	"fmt"

	"rag/internal/utils"
)

type PreviewDocumentUsecase struct {
	pdfProcessor PDFProcessor
	urlProcessor URLProcessor
}

func NewPreviewDocumentUsecase(
	pdfProcessor PDFProcessor,
	urlProcessor URLProcessor,
) *PreviewDocumentUsecase {
	return &PreviewDocumentUsecase{
		pdfProcessor: pdfProcessor,
		urlProcessor: urlProcessor,
	}
}

type PreviewResult struct {
	ExtractedText  string
	PagesExtracted int32
}

func (u *PreviewDocumentUsecase) Preview(ctx context.Context, domain *utils.PreviewDocumentDomain) (*utils.PreviewResult, error) {
	switch domain.SourceType {
	case utils.DocumentSourceTypeText:
		text, err := base64.StdEncoding.DecodeString(domain.ContentBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode text from base64: %w", err)
		}
		return &utils.PreviewResult{
			ExtractedText:  string(text),
			PagesExtracted: 0,
		}, nil

	case utils.DocumentSourceTypeURL:
		if u.urlProcessor == nil {
			return nil, fmt.Errorf("url processor not configured")
		}
		text, err := u.urlProcessor.ExtractTextFromURL(domain.SourceURL, domain.URLMaxDepth)
		if err != nil {
			return nil, fmt.Errorf("urlProcessor.ExtractTextFromURL got error: %w", err)
		}
		return &utils.PreviewResult{
			ExtractedText:  text,
			PagesExtracted: 0, // URL doesn't have pages
		}, nil

	case utils.DocumentSourceTypePDF:
		if u.pdfProcessor == nil {
			return nil, fmt.Errorf("pdf processor not configured")
		}
		pdfData, err := base64.StdEncoding.DecodeString(domain.ContentBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode PDF from base64: %w", err)
		}
		if len(pdfData) > 20*1024*1024 {
			return nil, fmt.Errorf("pdf size exceeds 20MB limit")
		}
		text, err := u.pdfProcessor.ExtractTextFromPDF(pdfData)
		if err != nil {
			return nil, fmt.Errorf("pdfProcessor.ExtractTextFromPDF got error: %w", err)
		}
		// TODO: get actual page count from kug_scrap output
		return &utils.PreviewResult{
			ExtractedText:  text,
			PagesExtracted: 0,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported source type for preview: %s", domain.SourceType)
	}
}
