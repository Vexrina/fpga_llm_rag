package processors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type LinkScraperProcessor struct {
	pythonPath string
	scriptPath string
	timeout    time.Duration
	cacheDir   string
}

func NewLinkScraperProcessor(pythonPath, scriptPath, cacheDir string, timeout time.Duration) *LinkScraperProcessor {
	return &LinkScraperProcessor{
		pythonPath: pythonPath,
		scriptPath: scriptPath,
		timeout:    timeout,
		cacheDir:   cacheDir,
	}
}

type DiscoverResult struct {
	Links []string `json:"links"`
}

func (p *LinkScraperProcessor) DiscoverLinks(url string, maxDepth int32) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "link-discovery-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "output.json")

	args := []string{
		p.scriptPath,
		"--url", url,
		"--max-depth", fmt.Sprintf("%d", maxDepth),
		"--links-only",
		"--output", outputPath,
		"--cache-dir", p.cacheDir,
	}

	cmd := exec.CommandContext(ctx, p.pythonPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("link scraper timeout after %v", p.timeout)
		}
		return nil, fmt.Errorf("link scraper failed: %w, stderr: %s", err, stderr.String())
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read output: %w", err)
	}

	var result struct {
		Links []string `json:"links"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}

	return result.Links, nil
}

func (p *LinkScraperProcessor) ScrapeTextFromURLs(urls []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, url := range urls {
		text, err := p.ExtractTextFromURL(url, 0)
		if err != nil {
			result[url] = fmt.Sprintf("Error: %v", err)
			continue
		}
		result[url] = text
	}

	return result, nil
}

func (p *LinkScraperProcessor) ExtractTextFromURL(url string, maxDepth int32) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	outputPath := filepath.Join(p.cacheDir, "output.json")

	args := []string{
		p.scriptPath,
		"--url", url,
		"--max-depth", fmt.Sprintf("%d", maxDepth),
		"--output", outputPath,
		"--cache-dir", p.cacheDir,
	}

	cmd := exec.CommandContext(ctx, p.pythonPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("url scraper timeout after %v", p.timeout)
		}
		return "", fmt.Errorf("url scraper failed: %w, stderr: %s", err, stderr.String())
	}

	extractOutput := filepath.Join(p.cacheDir, "texts.json")
	extractArgs := []string{
		p.scriptPath,
		"--url", url,
		"--cache-extract-text",
		"--cache-text-output", extractOutput,
		"--cache-dir", p.cacheDir,
	}

	cmd2 := exec.CommandContext(ctx, p.pythonPath, extractArgs...)
	cmd2.Stderr = &stderr

	if err := cmd2.Run(); err != nil {
		return "", fmt.Errorf("extract from cache failed: %w, stderr: %s", err, stderr.String())
	}

	extractData, err := os.ReadFile(extractOutput)
	if err != nil {
		return "", fmt.Errorf("failed to read extract output: %w", err)
	}

	var texts []struct {
		URL  string `json:"url"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(extractData, &texts); err != nil {
		return "", fmt.Errorf("failed to parse extract: %w", err)
	}

	for _, t := range texts {
		if t.URL == url || t.URL == url+"/" || url == t.URL+"/" {
			return t.Text, nil
		}
	}

	if len(texts) > 0 {
		return texts[0].Text, nil
	}

	return "", nil
}
