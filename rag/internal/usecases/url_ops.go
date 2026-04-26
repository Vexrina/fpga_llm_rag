package usecases

import (
	"context"
	"sync"
)

type DiscoverLinksUsecase struct {
	urlProcessor URLProcessor
}

func NewDiscoverLinksUsecase(urlProcessor URLProcessor) *DiscoverLinksUsecase {
	return &DiscoverLinksUsecase{
		urlProcessor: urlProcessor,
	}
}

func (u *DiscoverLinksUsecase) Discover(ctx context.Context, url string, maxDepth int32) ([]string, error) {
	if u.urlProcessor == nil {
		return nil, nil
	}
	return u.urlProcessor.DiscoverLinks(url, maxDepth)
}

type ScrapeUrlsUsecase struct {
	urlProcessor URLProcessor
}

func NewScrapeUrlsUsecase(urlProcessor URLProcessor) *ScrapeUrlsUsecase {
	return &ScrapeUrlsUsecase{
		urlProcessor: urlProcessor,
	}
}

func (u *ScrapeUrlsUsecase) Scrape(ctx context.Context, urls []string) (map[string]string, error) {
	if u.urlProcessor == nil {
		return nil, nil
	}

	result := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	for _, url := range urls {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			text, err := u.urlProcessor.ExtractTextFromURL(link, 0)
			if err != nil {
				once.Do(func() { firstErr = err })
				return
			}
			mu.Lock()
			result[link] = text
			mu.Unlock()
		}(url)
	}
	wg.Wait()

	if firstErr != nil {
		return result, firstErr
	}
	return result, nil
}

type URLTextEntry struct {
	URL  string
	Text string
}

func ToURLTextEntries(m map[string]string) []URLTextEntry {
	entries := make([]URLTextEntry, 0, len(m))
	for url, text := range m {
		entries = append(entries, URLTextEntry{
			URL:  url,
			Text: text,
		})
	}
	return entries
}
