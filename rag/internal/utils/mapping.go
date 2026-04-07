package utils

import pb "rag/pkg/rag"

func documentSourceTypeFromPB(t pb.DocumentSourceType) DocumentSourceType {
	switch t {
	case pb.DocumentSourceType_SOURCE_TYPE_TEXT:
		return DocumentSourceTypeText
	case pb.DocumentSourceType_SOURCE_TYPE_URL:
		return DocumentSourceTypeURL
	case pb.DocumentSourceType_SOURCE_TYPE_PDF:
		return DocumentSourceTypePDF
	default:
		return DocumentSourceTypeUnspecified
	}
}

func AddDocumentFromGRPCToDomain(req *pb.AddDocumentRequest) *AddDocumentDomain {
	return &AddDocumentDomain{
		Content:     req.GetContent(),
		Title:       req.GetTitle(),
		Metadata:    req.GetMetadata(),
		Embedding:   req.GetEmbedding(),
		SourceType:  documentSourceTypeFromPB(req.GetSourceType()),
		SourceURL:   req.GetSourceUrl(),
		URLMaxDepth: req.GetUrlMaxDepth(),
	}
}

func PreviewDocumentFromGRPCToDomain(req *pb.PreviewDocumentRequest) *PreviewDocumentDomain {
	return &PreviewDocumentDomain{
		Title:         req.GetTitle(),
		SourceType:    documentSourceTypeFromPB(req.GetSourceType()),
		SourceURL:     req.GetSourceUrl(),
		ContentBase64: req.GetContentBase64(),
		URLMaxDepth:   req.GetUrlMaxDepth(),
	}
}

func PreviewResultToDomainToGRPC(res *PreviewResult) *pb.PreviewDocumentResponse {
	return &pb.PreviewDocumentResponse{
		ExtractedText:  res.ExtractedText,
		PagesExtracted: res.PagesExtracted,
	}
}

func CommitDocumentFromGRPCToDomain(req *pb.CommitDocumentRequest) *CommitDocumentDomain {
	return &CommitDocumentDomain{
		Title:    req.GetTitle(),
		Content:  req.GetContent(),
		Metadata: req.GetMetadata(),
	}
}

func CommitResultToGRPC(id string) *pb.CommitDocumentResponse {
	return &pb.CommitDocumentResponse{
		Success: true,
		Message: "Document committed successfully",
		Id:      id,
	}
}

func AddDocumentFromDomainToGRPC() *pb.AddDocumentResponse {
	return &pb.AddDocumentResponse{} // TODO: IMPLEMENT ME
}

func GetDocumentFromGRPCToDomain(req *pb.GetDocumentRequest) *GetDocumentDomain {
	return &GetDocumentDomain{
		Id: req.GetId(),
	}
}

func GetDocumentFromDomainToGRPC() *pb.GetDocumentResponse {
	return &pb.GetDocumentResponse{} // TODO: IMPLEMENT ME
}

func DeleteDocumentFromGRPCToDomain(req *pb.DeleteDocumentRequest) *DeleteDocumentDomain {
	return &DeleteDocumentDomain{
		Id: req.GetId(),
	}
}

func DeleteDocumentFromDomainToGRPC() *pb.DeleteDocumentResponse {
	return &pb.DeleteDocumentResponse{} // TODO: IMPLEMENT ME
}

func SearchDocumentFromGRPCToDomain(req *pb.SearchRequest) *SearchDocumentDomain {

	return &SearchDocumentDomain{
		Query:         req.GetQuery(),
		Limit:         req.GetLimit(),
		SimilarityThs: req.GetSimilarityThreshold(),
	}
}

func SearchDocumentFromDomainToGRPC() *pb.SearchResponse {
	return &pb.SearchResponse{} // TODO: IMPLEMENT ME
}
