package app

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "rag/pkg/rag"
)

func (s *RagServer) DiscoverLinks(ctx context.Context, req *pb.DiscoverLinksRequest) (*pb.DiscoverLinksResponse, error) {
	if req.GetUrl() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "url is required")
	}

	links, err := s.discoverLinksUsecase.Discover(ctx, req.GetUrl(), req.GetMaxDepth())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "discoverLinksUsecase.Discover error: %s", err)
	}

	return &pb.DiscoverLinksResponse{
		Links: links,
	}, nil
}

func (s *RagServer) ScrapeUrls(ctx context.Context, req *pb.ScrapeUrlsRequest) (*pb.ScrapeUrlsResponse, error) {
	if len(req.GetUrls()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "urls are required")
	}

	texts, err := s.scrapeUrlsUsecase.Scrape(ctx, req.GetUrls())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "scrapeUrlsUsecase.Scrape error: %s", err)
	}

	result := make([]*pb.ScrapedTextEntry, 0, len(texts))
	for url, text := range texts {
		result = append(result, &pb.ScrapedTextEntry{
			Url:  url,
			Text: text,
		})
	}

	return &pb.ScrapeUrlsResponse{
		Texts: result,
	}, nil
}
