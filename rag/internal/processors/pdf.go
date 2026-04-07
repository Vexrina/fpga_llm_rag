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

	args := append([]string{p.pythonPath, p.scriptPath}, p.pythonArgs...)
	args = append(args, "--pdf", inputPath)
	args = append(args, "--pages", "1") // all pages - will extract all
	args = append(args, "-o", outputPath)

	cmd := exec.CommandContext(ctx, "python3", args...)
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
