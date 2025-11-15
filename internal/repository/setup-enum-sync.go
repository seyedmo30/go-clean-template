package repository

import (
	"fmt"
	"slices"
	"strings"

	"gorm.io/gorm"
)

// getCurrentEnumValues reads current enum values from MySQL INFORMATION_SCHEMA.
func getCurrentEnumValues(db *gorm.DB, tableName, columnName string) ([]string, error) {
	var columnType string
	err := db.
		Raw(`
			SELECT COLUMN_TYPE
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND COLUMN_NAME = ?
		`, tableName, columnName).
		Row().
		Scan(&columnType)
	if err != nil {
		return nil, fmt.Errorf("failed to read enum column type: %w", err)
	}

	// Extract values from enum('A','B',...) format
	inside := strings.TrimPrefix(columnType, "enum(")
	inside = strings.TrimSuffix(inside, ")")
	parts := strings.Split(inside, ",")
	for i, p := range parts {
		parts[i] = strings.Trim(p, `'`)
	}
	return parts, nil
}

// SyncEnumColumn ensures the ENUM column includes all values in the given map[int]string.
func SyncEnumColumn(db *gorm.DB, tableName, columnName string, enumMap map[int]string) error {
	currentValues, err := getCurrentEnumValues(db, tableName, columnName)
	if err != nil {
		return err
	}

	// Build a unique union of current and enum values
	desiredValues := make([]string, 0, len(enumMap))
	for _, v := range enumMap {
		desiredValues = append(desiredValues, v)
	}

	// Sort and remove duplicates
	slices.Sort(desiredValues)
	slices.Sort(currentValues)

	if slices.Equal(desiredValues, currentValues) {
		return nil // Already up-to-date
	}

	// Generate ALTER TABLE statement
	escaped := make([]string, 0, len(desiredValues))
	for _, v := range desiredValues {
		escaped = append(escaped, fmt.Sprintf("'%s'", v))
	}

	sql := fmt.Sprintf(`ALTER TABLE %s MODIFY %s ENUM(%s)`, tableName, columnName, strings.Join(escaped, ","))
	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to alter enum column: %w", err)
	}

	return nil
}

func convertEnumMap[E ~int](enumMap map[E]string) map[int]string {
	result := make(map[int]string, len(enumMap))
	for k, v := range enumMap {
		result[int(k)] = v
	}
	return result
}
