package repository

import (
	"database/sql"
	"fmt"
)

// Пример использования векторной обертки с Jet
func ExampleVectorUsage(db *sql.DB) error {
	// Предполагаем, что у нас есть сгенерированные Jet модели
	// table := jetdb.Documents

	// 1. Вставка документа с вектором
	embedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	vectorValue := VectorFromFloat32(embedding)

	// Пример SQL запроса для вставки (замените на реальные Jet модели)
	insertQuery := `
		INSERT INTO documents (title, content, metadata, embedding)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(insertQuery,
		"Пример документа",
		"Содержание документа",
		`{"source": "example"}`,
		vectorValue)
	if err != nil {
		return fmt.Errorf("failed to insert document: %v", err)
	}

	// 2. Поиск похожих документов
	queryEmbedding := []float32{0.11, 0.21, 0.31, 0.41, 0.51}
	queryVectorValue := VectorFromFloat32(queryEmbedding)

	// SQL запрос для поиска похожих документов
	searchQuery := `
		SELECT id, title, content, metadata, embedding,
		       embedding <=> $1 as distance
		FROM documents
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT 5
	`

	rows, err := db.Query(searchQuery, queryVectorValue)
	if err != nil {
		return fmt.Errorf("failed to search documents: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title, content string
		var metadata []byte
		var vectorScan VectorScan
		var distance float64

		err := rows.Scan(&id, &title, &content, &metadata, &vectorScan, &distance)
		if err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		fmt.Printf("Document ID: %d, Title: %s, Distance: %f\n", id, title, distance)
		fmt.Printf("Vector: %v\n", vectorScan.Vector)
	}

	return nil
}

// Пример использования с nullable векторами
func ExampleNullableVectorUsage(db *sql.DB) error {
	// Вставка документа с nullable вектором
	embedding := []float32{0.1, 0.2, 0.3}
	vectorValueNull := VectorFromFloat32Null(embedding, true)

	insertQuery := `
		INSERT INTO documents (title, content, embedding)
		VALUES ($1, $2, $3)
	`

	_, err := db.Exec(insertQuery,
		"Документ с вектором",
		"Содержание",
		vectorValueNull)
	if err != nil {
		return fmt.Errorf("failed to insert document: %v", err)
	}

	// Чтение документа с nullable вектором
	selectQuery := `
		SELECT id, title, embedding
		FROM documents
		WHERE title = $1
	`

	var id int
	var title string
	var vectorScanNull VectorScanNull

	err = db.QueryRow(selectQuery, "Документ с вектором").Scan(&id, &title, &vectorScanNull)
	if err != nil {
		return fmt.Errorf("failed to select document: %v", err)
	}

	if vectorScanNull.Valid {
		fmt.Printf("Document ID: %d, Title: %s\n", id, title)
		fmt.Printf("Vector: %v\n", vectorScanNull.Vector)
	} else {
		fmt.Printf("Document ID: %d, Title: %s (no vector)\n", id, title)
	}

	return nil
}

// Пример использования SQL функций для векторов
func ExampleVectorSQLFunctions() {
	// Генерация SQL выражений для различных операций с векторами

	// Косинусное расстояние
	cosineDistanceSQL := VectorCosineDistanceSQL("documents.embedding", "$1")
	fmt.Printf("Cosine distance SQL: %s\n", cosineDistanceSQL)

	// L2 расстояние
	l2DistanceSQL := VectorL2DistanceSQL("documents.embedding", "$1")
	fmt.Printf("L2 distance SQL: %s\n", l2DistanceSQL)

	// Косинусное сходство
	similaritySQL := VectorSimilaritySQL("documents.embedding", "$1")
	fmt.Printf("Similarity SQL: %s\n", similaritySQL)

	// Внутреннее произведение
	innerProductSQL := VectorInnerProductSQL("documents.embedding", "$1")
	fmt.Printf("Inner product SQL: %s\n", innerProductSQL)
}
