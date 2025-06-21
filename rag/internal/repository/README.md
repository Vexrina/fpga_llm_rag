# Векторная обертка для Jet и PostgreSQL

Этот пакет предоставляет обертку для работы с векторными типами данных (pgvector) в PostgreSQL через Jet ORM.

## Основные компоненты

### VectorScan
Структура для сканирования векторов из базы данных:
```go
type VectorScan struct {
    Vector []float32
}
```

### VectorValue
Структура для вставки векторов в базу данных:
```go
type VectorValue struct {
    Vector []float32
}
```

### VectorScanNull / VectorValueNull
Аналогичные структуры для работы с nullable векторами.

## Использование

### 1. Вставка векторов

```go
// Создание вектора из []float32
embedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
vectorValue := VectorFromFloat32(embedding)

// Вставка в базу данных
_, err := db.Exec(`
    INSERT INTO documents (title, content, embedding)
    VALUES ($1, $2, $3)
`, "Заголовок", "Содержание", vectorValue)
```

### 2. Чтение векторов

```go
// Сканирование вектора из результата запроса
var vectorScan VectorScan
err := db.QueryRow(`
    SELECT embedding FROM documents WHERE id = $1
`, documentID).Scan(&vectorScan)

if err != nil {
    return err
}

// Получение []float32 из VectorScan
vector := VectorToFloat32(vectorScan)
fmt.Printf("Vector: %v\n", vector)
```

### 3. Поиск похожих документов

```go
// Создание вектора запроса
queryEmbedding := []float32{0.11, 0.21, 0.31, 0.41, 0.51}
queryVectorValue := VectorFromFloat32(queryEmbedding)

// Поиск похожих документов
rows, err := db.Query(`
    SELECT id, title, embedding <=> $1 as distance
    FROM documents
    WHERE embedding IS NOT NULL
    ORDER BY embedding <=> $1
    LIMIT 5
`, queryVectorValue)

for rows.Next() {
    var id int
    var title string
    var distance float64
    var vectorScan VectorScan
    
    err := rows.Scan(&id, &title, &distance, &vectorScan)
    if err != nil {
        return err
    }
    
    fmt.Printf("ID: %d, Title: %s, Distance: %f\n", id, title, distance)
}
```

### 4. Работа с nullable векторами

```go
// Вставка nullable вектора
embedding := []float32{0.1, 0.2, 0.3}
vectorValueNull := VectorFromFloat32Null(embedding, true)

_, err := db.Exec(`
    INSERT INTO documents (title, embedding)
    VALUES ($1, $2)
`, "Документ", vectorValueNull)

// Чтение nullable вектора
var vectorScanNull VectorScanNull
err := db.QueryRow(`
    SELECT embedding FROM documents WHERE id = $1
`, documentID).Scan(&vectorScanNull)

if vectorScanNull.Valid {
    fmt.Printf("Vector: %v\n", vectorScanNull.Vector)
} else {
    fmt.Println("No vector")
}
```

## SQL функции для векторов

Пакет предоставляет функции для генерации SQL выражений:

```go
// Косинусное расстояние
cosineDistanceSQL := VectorCosineDistanceSQL("documents.embedding", "$1")
// Результат: "documents.embedding <-> $1"

// L2 расстояние
l2DistanceSQL := VectorL2DistanceSQL("documents.embedding", "$1")
// Результат: "documents.embedding <-> $1"

// Косинусное сходство
similaritySQL := VectorSimilaritySQL("documents.embedding", "$1")
// Результат: "1 - (documents.embedding <-> $1)"

// Внутреннее произведение
innerProductSQL := VectorInnerProductSQL("documents.embedding", "$1")
// Результат: "documents.embedding <#> $1"

// Общее расстояние
distanceSQL := VectorDistanceSQL("documents.embedding", "$1")
// Результат: "documents.embedding <=> $1"
```

## Интеграция с Jet

После генерации схем Jet, вы можете использовать векторную обертку следующим образом:

```go
// Предполагаем, что у нас есть сгенерированные Jet модели
// table := jetdb.Documents

// Вставка с использованием Jet
insertQuery := table.INSERT(table.Title, table.Content, table.Embedding).
    VALUES("Заголовок", "Содержание", VectorFromFloat32([]float32{0.1, 0.2, 0.3}))

// Поиск с использованием Jet
searchQuery := table.SELECT(table.ID, table.Title, table.Embedding).
    WHERE(table.Embedding.IS_NOT_NULL()).
    ORDER_BY(VectorDistanceSQL("documents.embedding", "$1")).
    LIMIT(5)
```

## Примечания

1. **Формат векторов**: Векторы хранятся в PostgreSQL в формате `[1.0,2.0,3.0]`
2. **Размерность**: Убедитесь, что все векторы имеют одинаковую размерность
3. **Индексы**: Для эффективного поиска создайте индекс на поле embedding:
   ```sql
   CREATE INDEX ON documents USING ivfflat (embedding vector_cosine_ops);
   ```

## Примеры

Смотрите файл `vector_example.go` для полных примеров использования. 