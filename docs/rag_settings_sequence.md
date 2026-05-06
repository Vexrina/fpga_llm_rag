```mermaid
sequenceDiagram
    actor u as User
    participant react as React Frontend
    participant proxy as "Vite Proxy"
    participant gql as "GraphQL Gateway"
    participant rag as "RAG Service"
    participant db as PostgreSQL
    participant llm as "LLM Gateway"
    participant ollama as Ollama

    u->>react: Открыть "Настройки RAG"

    rect rgb(200, 220, 255)
        Note over react: Загрузка текущих настроек
        react->>proxy: query getRagSettings
        proxy->>gql: HTTP POST /graphql
        gql->>rag: gRPC GetRagSettings()
        rag->>db: SELECT * FROM rag_settings
        db->>rag: { topK, chunkSize, basePrompt, ... }
        rag->>gql: { settings: [...] }
        gql->>proxy: GraphQL response
        proxy->>react: Показать настройки
    end

    u->>react: Изменить настройку<br/>(например, topK)

    react->>react: update state

    u->>react: Нажать "Сохранить"

    rect rgb(200, 255, 200)
        Note over react, rag: Сохранение настройки
        react->>proxy: mutation updateRagSetting(key, value)
        proxy->>gql: HTTP POST /graphql
        gql->>rag: gRPC UpdateRagSettings()

        Note over rag: Валидация значения
        alt Невалидное значение
            rag->>gql: error: "invalid integer value"
            gql->>proxy: { errors: [...] }
            proxy->>react: Показать ошибку
            react->>u: Показать ошибку валидации
        else Валидное значение
            rag->>db: INSERT INTO rag_settings_history<br/>(key, old_value, new_value, changed_by)
            rag->>db: UPDATE rag_settings SET value = $value
            db->>rag: OK

            alt key == "basePrompt"
                Note over rag, llm: Уведомление LLM Gateway
                rag->>llm: gRPC UpdateBasePrompt(prompt)
                
                rect rgb(255, 200, 200)
                    Note over llm: Обновление базового промпта<br/>AskUsecase.basePrompt
                end

                llm->>llm: AskUsecase.UpdateBasePrompt()
                Note over llm: Следующий Ask() использует<br/>новый basePrompt
            end

            rag->>gql: { success: true }
            gql->>proxy: GraphQL response
            proxy->>react: Настройка сохранена
            react->>u: Показать "Сохранено"
        end
    end
```

## Валидация настроек в RAG Service

| Параметр | Тип | Валидация |
|----------|-----|-----------|
| `topK` | int | Должно быть числом |
| `chunkSize` | int | Должно быть числом |
| `chunkOverlap` | int | Должно быть числом |
| `similarityThreshold` | float | Должно быть float |
| `comparisonMethod` | string | one of: cosine, dot, euclidean, l1 |
| `basePrompt` | string | **Должен содержать** "контекст" и "вопрос" (case insensitive) |

## Особенность basePrompt

При изменении `basePrompt` RAG Service **асинхронно** уведомляет LLM Gateway:
- RAG → LLM Gateway: `gRPC UpdateBasePrompt(prompt)`
- LLM Gateway → `AskUsecase.UpdateBasePrompt()` 
- Следующие запросы `Ask()` будут использовать новый промпт

```mermaid
sequenceDiagram
    participant rag as "RAG Service"
    participant llm as "LLM Gateway"
    rag->>llm: async UpdateBasePrompt()
    Note over llm: AskUsecase.basePrompt = newValue
    Note over llm: (не блокирует ответ)
```