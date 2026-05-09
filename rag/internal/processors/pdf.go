package processors

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type KugScrapProcessor struct {
	pythonPath string
	scriptPath string
	pythonArgs []string
	timeout    time.Duration
}

func NewKugScrapProcessor(pythonPath, scriptPath string, pythonArgs []string, timeout time.Duration) *KugScrapProcessor {
	return &KugScrapProcessor{
		pythonPath: pythonPath,
		scriptPath: scriptPath,
		pythonArgs: pythonArgs,
		timeout:    timeout,
	}
}

func (p *KugScrapProcessor) ExtractTextFromPDF(pdfData []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "pdf-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := fmt.Sprintf("input_%s.pdf", uuid.New().String()[:8])
	inputPath := filepath.Join(tmpDir, filename)
	outputPath := filepath.Join(tmpDir, "output.txt")

	if err := os.WriteFile(inputPath, pdfData, 0644); err != nil {
		return "", fmt.Errorf("failed to write PDF: %w", err)
	}

	pageCount, err := p.countPdfPages(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to count PDF pages: %w", err)
	}

	args := append([]string{p.scriptPath}, p.pythonArgs...)
	args = append(args, "--pdf", inputPath)

	pages := make([]string, 0, pageCount)
	for i := 1; i <= pageCount; i++ {
		pages = append(pages, fmt.Sprintf("%d", i))
	}
	args = append(args, "--pages")
	args = append(args, pages...)

	args = append(args, "--dpi", "200")
	args = append(args, "--strategy", "auto")
	args = append(args, "--out", outputPath)

	cmd := exec.CommandContext(ctx, p.pythonPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("kug_scrap timeout after %v", p.timeout)
		}
		return "", fmt.Errorf("kug_scrap failed: %w, stderr: %s", err, stderr.String())
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}

	text := strings.TrimSpace(string(output))
	if text == "" {
		return "", fmt.Errorf("kug_scrap produced empty output")
	}

	return text, nil
}

func (p *KugScrapProcessor) countPdfPages(pdfPath string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.pythonPath, "-c",
		"import pdfplumber; pdf = pdfplumber.open('"+pdfPath+"'); print(len(pdf.pages))",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to count pages: %w, stderr: %s", err, stderr.String())
	}

	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(stdout.String()), "%d", &count); err != nil {
		return 0, fmt.Errorf("failed to parse page count: %w, output: %s", err, stdout.String())
	}

	return count, nil
}
