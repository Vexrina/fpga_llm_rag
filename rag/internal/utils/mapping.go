package utils

import pb "rag/pkg/rag"

func AddDocumentFromGRPCToDomain(req *pb.AddDocumentRequest) *AddDocumentDomain {
	return &AddDocumentDomain{
		Content:   req.GetContent(),
		Title:     req.GetTitle(),
		Metadata:  req.GetMetadata(),
		Embedding: req.GetEmbedding(),
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
