```mermaid
sequenceDiagram
    actor u as User
    participant react as React Frontend
    participant proxy as Vite Proxy
    participant gql as GraphQL Gateway
    participant rag as RAG Service
    participant db as PostgreSQL<br/>(pgvector)
    participant fw as Float-Weaver
    participant tgi as TGI<br/>(Embeddings)
    participant py as Python<br/>(scraper/pdf)

    u->>react: Нажми "Добавить документ"

    rect rgb(200, 220, 255)
        Note over react: Выбор типа: URL / TEXT / PDF
    end

    alt URL Source Type
        rect rgb(255, 240, 200)
            Note over react, rag: ШАГ 1: Discover Links
            react->>proxy: mutation discoverLinks(url, maxDepth)
            proxy->>gql: HTTP POST /graphql
            gql->>rag: gRPC DiscoverLinks()
            rag->>py: exec link_scraper_cached.py
            Note over py: Запуск браузера<br/>Обход ссылок до maxDepth
            py->>rag: []links
            rag->>gql: { links: [...] }
            gql->>proxy: GraphQL response
            proxy->>react: Показать найденные ссылки
        end

        rect rgb(255, 240, 200)
            Note over react, rag: ШАГ 2: Scrape URLs
            react->>proxy: mutation scrapeUrls(urls[])
            proxy->>gql: HTTP POST /graphql
            gql->>rag: gRPC ScrapeUrls()
            rag->>py: exec link_scraper_cached.py --cache-extract-text
            Note over py: Извлечение текста<br/>со страниц
            py->>rag: { url: text }
            rag->>gql: { texts: [...] }
            gql->>proxy: GraphQL response
            proxy->>react: Показать scraped текст
            Note over react: Пользователь редактирует<br/>название и содержимое
        end
    end

    alt PDF Source Type
        rect rgb(200, 200, 255)
            Note over react: Конвертация файла<br/>в base64
            react->>react: FileReader.readAsDataURL()
        end
    end

    rect rgb(200, 255, 200)
        Note over react, rag: ШАГ 3: Commit Document
        react->>proxy: mutation commitDocument(title, content)
        proxy->>gql: HTTP POST /graphql
        gql->>rag: gRPC CommitDocument()

        Note over rag: Разбиение текста на чанки<br/>chunkText() - maxTokens=200
        rag->>rag: Разбить content на chunks

        loop Для каждого чанка
            rag->>fw: gRPC Embed(text: chunk)
            fw->>tgi: HTTP POST /embed
            
            rect rgb(200, 200, 200)
                Note over tgi: Генерация эмбеддинга<br/>BAAI/bge-large-en-v1.5
            end
            
            tgi->>fw: embedding vector
            fw->>rag: Embeddings[0].Values

            Note over rag: Вставка в PostgreSQL
            rag->>db: INSERT documents (title, embedding, text, metadata)
            db->>rag: OK
        end

        rag->>gql: { success: true, id: "..." }
        gql->>proxy: GraphQL response
        proxy->>react: Документ сохранён
    end

    react->>u: Показать документ в списке

    rect rgb(200, 220, 255)
        Note over rag, db: Документ готов для RAG поиска
    end
```

## Варианты добавления

### 1. URL
```
User → discoverLinks → scrapeUrls → commitDocument
```

### 2. TEXT  
```
User → commitDocument (напрямую с текстом)
```

### 3. PDF
```
User → base64 encode → commitDocument
```

## Ключевые операции в RAG Service:

| Операция | Описание |
|----------|----------|
| `discoverLinks` | Обход страницы, поиск всех ссылок до maxDepth |
| `scrapeUrls` | Извлечение текста с выбранных URL |
| `commitDocument` | 1) Разбить на чанки (200 токенов) → 2) Получить эмбеддинг для каждого → 3) Сохранить в pgvector |