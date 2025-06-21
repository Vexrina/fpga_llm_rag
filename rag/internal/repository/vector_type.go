package repository

import (
	"fmt"
	"strings"
)

// VectorScan реализует интерфейс для сканирования векторов из базы данных
type VectorScan struct {
	Vector []float32
}

// Scan реализует интерфейс sql.Scanner для векторов
func (v *VectorScan) Scan(value interface{}) error {
	if value == nil {
		v.Vector = nil
		return nil
	}

	switch val := value.(type) {
	case []byte:
		// Предполагаем, что вектор приходит как строка в формате [1.0,2.0,3.0]
		str := string(val)
		str = strings.Trim(str, "[]")
		if str == "" {
			v.Vector = []float32{}
			return nil
		}

		parts := strings.Split(str, ",")
		vector := make([]float32, len(parts))
		for i, part := range parts {
			var f float64
			_, err := fmt.Sscanf(strings.TrimSpace(part), "%f", &f)
			if err != nil {
				return fmt.Errorf("failed to parse vector value: %v", err)
			}
			vector[i] = float32(f)
		}
		v.Vector = vector
		return nil
	default:
		return fmt.Errorf("cannot scan %T into VectorScan", value)
	}
}

// VectorValue реализует интерфейс для вставки векторов в базу данных
type VectorValue struct {
	Vector []float32
}

// String возвращает строковое представление вектора для Jet
func (v VectorValue) String() string {
	if v.Vector == nil {
		return "NULL"
	}

	values := make([]string, len(v.Vector))
	for i, val := range v.Vector {
		values[i] = fmt.Sprintf("%f", val)
	}
	return fmt.Sprintf("[%s]", strings.Join(values, ","))
}

// Value реализует интерфейс driver.Valuer для векторов
func (v VectorValue) Value() (interface{}, error) {
	if v.Vector == nil {
		return nil, nil
	}

	values := make([]string, len(v.Vector))
	for i, val := range v.Vector {
		values[i] = fmt.Sprintf("%f", val)
	}
	return fmt.Sprintf("[%s]", strings.Join(values, ",")), nil
}

// VectorFromFloat32 создает VectorValue из слайса float32
func VectorFromFloat32(vector []float32) VectorValue {
	return VectorValue{Vector: vector}
}

// VectorToFloat32 конвертирует VectorScan в слайс float32
func VectorToFloat32(scan VectorScan) []float32 {
	return scan.Vector
}

// VectorDistanceSQL возвращает SQL выражение для вычисления расстояния между векторами
func VectorDistanceSQL(column1, column2 string) string {
	return fmt.Sprintf("%s <=> %s", column1, column2)
}

// VectorCosineDistanceSQL возвращает SQL выражение для вычисления косинусного расстояния
func VectorCosineDistanceSQL(column1, column2 string) string {
	return fmt.Sprintf("%s <-> %s", column1, column2)
}

// VectorL2DistanceSQL возвращает SQL выражение для вычисления L2 расстояния
func VectorL2DistanceSQL(column1, column2 string) string {
	return fmt.Sprintf("%s <-> %s", column1, column2)
}

// VectorInnerProductSQL возвращает SQL выражение для вычисления внутреннего произведения
func VectorInnerProductSQL(column1, column2 string) string {
	return fmt.Sprintf("%s <#> %s", column1, column2)
}

// VectorSimilaritySQL возвращает SQL выражение для вычисления косинусного сходства
func VectorSimilaritySQL(column1, column2 string) string {
	return fmt.Sprintf("1 - (%s <-> %s)", column1, column2)
}

// VectorScanNull реализует интерфейс для сканирования nullable векторов
type VectorScanNull struct {
	Vector []float32
	Valid  bool
}

// Scan реализует интерфейс sql.Scanner для nullable векторов
func (v *VectorScanNull) Scan(value interface{}) error {
	if value == nil {
		v.Vector = nil
		v.Valid = false
		return nil
	}

	v.Valid = true
	scan := VectorScan{Vector: v.Vector}
	return scan.Scan(value)
}

// VectorValueNull реализует интерфейс для вставки nullable векторов
type VectorValueNull struct {
	Vector []float32
	Valid  bool
}

// Value реализует интерфейс driver.Valuer для nullable векторов
func (v VectorValueNull) Value() (interface{}, error) {
	if !v.Valid {
		return nil, nil
	}
	return VectorValue{Vector: v.Vector}.Value()
}

// VectorFromFloat32Null создает VectorValueNull из слайса float32
func VectorFromFloat32Null(vector []float32, valid bool) VectorValueNull {
	return VectorValueNull{Vector: vector, Valid: valid}
}
