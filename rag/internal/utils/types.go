package utils

type DocumentSourceType string

const (
	DocumentSourceTypeUnspecified DocumentSourceType = ""
	DocumentSourceTypeText        DocumentSourceType = "text"
	DocumentSourceTypeURL         DocumentSourceType = "url"
	DocumentSourceTypePDF         DocumentSourceType = "pdf"
)

type ComparisonMethod string

const (
	ComparisonMethodUnspecified ComparisonMethod = ""
	ComparisonMethodCosine      ComparisonMethod = "cosine"
	ComparisonMethodDot         ComparisonMethod = "dot"
	ComparisonMethodEuclidean   ComparisonMethod = "euclidean"
	ComparisonMethodL1          ComparisonMethod = "l1"
)

type (
	AddDocumentDomain struct {
		Id          string
		Content     string
		Title       string
		Metadata    map[string]string
		Embedding   []float32
		SourceType  DocumentSourceType
		SourceURL   string
		URLMaxDepth int32
	}

	PreviewDocumentDomain struct {
		Title         string
		SourceType    DocumentSourceType
		SourceURL     string
		ContentBase64 string
		URLMaxDepth   int32
	}

	CommitDocumentDomain struct {
		Title    string
		Content  string
		Metadata map[string]string
	}

	GetDocumentDomain struct {
		Id string
	}

	DeleteDocumentDomain struct {
		Id string
	}

	SearchDocumentDomain struct {
		Query            string
		Limit            int32
		SimilarityThs    float32
		ComparisonMethod ComparisonMethod
	}

	PreviewResult struct {
		ExtractedText  string
		PagesExtracted int32
	}

	RollbackDocumentDomain struct {
		DocumentID string
		VersionID  int32
		RollbackBy string
	}
)
