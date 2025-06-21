package utils

type (
	AddDocumentDomain struct {
		Id        string
		Content   string
		Title     string
		Metadata  map[string]string
		Embedding []float32
	}

	GetDocumentDomain struct {
		Id string
	}

	DeleteDocumentDomain struct {
		Id string
	}

	SearchDocumentDomain struct {
		Query         string
		Limit         int32
		SimilarityThs float32
	}
)
