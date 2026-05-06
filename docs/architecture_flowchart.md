```mermaid
flowchart LR
    subgraph Frontend
        UI["Frontend (React + Vite, :3000)"]
    end

    subgraph GraphQL
        GQL["GraphQL Gateway (:4000)"]
    end

    subgraph LLM_Gateway
        LLM["LLM Gateway (:8083)"]
    end

    subgraph RAG
        RAG["RAG Service (:50051)"]
    end

    subgraph DB
        PG["PostgreSQL (:5432)"]
    end

    subgraph Embedding
        FW["Float-Weaver (:8081)"]
        TGI["TGI (:80)"]
    end

    subgraph LLM_Model
        OLL["Ollama (:11434)"]
    end

    subgraph Python
        PY["Python Parsers"]
    end

    UI -->|"HTTP /graphql"| GQL
    GQL -->|"gRPC Ask"| LLM
    GQL -->|"gRPC Search"| RAG
    GQL -->|"gRPC Preview"| RAG
    GQL -->|"gRPC Commit"| RAG
    GQL -->|"gRPC Settings"| RAG

    LLM -->|"gRPC Search"| RAG
    LLM -->|"HTTP /api/generate"| OLL

    RAG -->|"gRPC Embed"| FW
    RAG -->|"SQL"| PG

    FW -->|"HTTP /embed"| TGI

    RAG -->|"exec"| PY
```

## Full Architecture (simplified)

```mermaid
flowchart TB
    FE["Frontend<br/>React + Vite<br/>:3000"]
    
    GQL["GraphQL Gateway<br/>gqlgen<br/>:4000"]
    
    LLM["LLM Gateway<br/>Go + gRPC<br/>:8083"]
    
    RAG["RAG Service<br/>Go + gRPC<br/>:50051"]
    
    PG["PostgreSQL<br/>pgvector<br/>:5432"]
    
    FW["Float-Weaver<br/>Go + gRPC<br/>:8081"]
    
    TGI["TGI<br/>text-embeddings-inference<br/>:80"]
    
    OLL["Ollama<br/>LLM Model<br/>:11434"]
    
    PY["Python<br/>Parsers<br/>scraper + PDF"]

    FE -->|"1 HTTP"| GQL
    GQL -->|"2 gRPC"| LLM
    GQL -->|"2 gRPC"| RAG
    
    LLM -->|"3 gRPC"| RAG
    LLM -->|"4 HTTP"| OLL
    
    RAG -->|"5 gRPC"| FW
    RAG -->|"6 SQL"| PG
    
    FW -->|"7 HTTP"| TGI
    
    RAG -->|"8 exec"| PY

    style FE fill:#e3f2fd,stroke:#1976d2
    style GQL fill:#e3f2fd,stroke:#1976d2
    style LLM fill:#e3f2fd,stroke:#1976d2
    style RAG fill:#e3f2fd,stroke:#1976d2
    style PG fill:#fff3e0,stroke:#f57c00
    style FW fill:#e3f2fd,stroke:#1976d2
    style TGI fill:#e8f5e9,stroke:#388e3c
    style OLL fill:#e8f5e9,stroke:#388e3c
    style PY fill:#fce4ec,stroke:#c2185b
```

## Operations Flow

### Ask Question
```mermaid
flowchart LR
    U[User] --> FE[Frontend]
    FE --> GQL[GraphQL]
    GQL --> LLM[LLM Gateway]
    LLM --> RAG[RAG]
    RAG --> FW[Float-Weaver]
    FW --> TGI[TGI]
    TGI --> FW
    FW --> RAG
    RAG --> PG[PostgreSQL]
    PG --> RAG
    RAG --> LLM
    LLM --> OLL[Ollama]
    OLL --> LLM
    LLM --> GQL
    GQL --> FE
    FE --> U
```

### Add Document (URL)
```mermaid
flowchart LR
    A[Admin] --> FE[Frontend]
    FE --> GQL[GraphQL]
    GQL --> RAG[RAG]
    RAG --> PY[Python Scraper]
    PY --> OCR[OCR]
    OCR --> PY
    PY --> RAG
    RAG --> FW[Float-Weaver]
    FW --> TGI[TGI]
    TGI --> FW
    FW --> RAG
    RAG --> PG[PostgreSQL]
```

### Update Settings
```mermaid
flowchart LR
    A[Admin] --> FE[Frontend]
    FE --> GQL[GraphQL]
    GQL --> RAG[RAG]
    RAG --> PG[PostgreSQL]
    PG --> RAG
    RAG -->|"if basePrompt"| LLM[LLM Gateway]
    LLM --> RAG
    RAG --> GQL
    GQL --> FE
    FE --> A
```