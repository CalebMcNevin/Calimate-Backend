package utils

import (
	"reflect"
	"strings"
	"time"
	"unicode"

	"gorm.io/gorm"
)

// ApplyFilter applies filter conditions to a GORM query using reflection
func ApplyFilter(query *gorm.DB, filter any) *gorm.DB {
	v := reflect.ValueOf(filter)
	t := reflect.TypeOf(filter)

	// Handle pointer to struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	// Skip if not a struct
	if v.Kind() != reflect.Struct {
		return query
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Skip nil pointers
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		// Get the actual value if it's a pointer
		var value any
		if field.Kind() == reflect.Ptr {
			value = field.Elem().Interface()
		} else {
			value = field.Interface()
		}

		// Get column name from tags or field name
		columnName := getColumnName(fieldType)

		// Apply appropriate filter based on field name and type
		fieldName := fieldType.Name
		switch v := value.(type) {
		case time.Time:
			if strings.Contains(fieldName, "From") || strings.HasSuffix(fieldName, "Start") {
				query = query.Where(columnName+" >= ?", v)
			} else if strings.Contains(fieldName, "To") || strings.HasSuffix(fieldName, "End") {
				query = query.Where(columnName+" <= ?", v)
			} else {
				query = query.Where(columnName+" = ?", v)
			}
		case SimpleDate:
			if strings.Contains(fieldName, "From") || strings.HasSuffix(fieldName, "Start") {
				// Start of day: 00:00:00
				startOfDay := v.Time
				query = query.Where(columnName+" >= ?", startOfDay)
			} else if strings.Contains(fieldName, "To") || strings.HasSuffix(fieldName, "End") {
				// End of day: 23:59:59.999
				endOfDay := v.Time.AddDate(0, 0, 1).Add(-1 * time.Nanosecond)
				query = query.Where(columnName+" <= ?", endOfDay)
			} else {
				// Exact date match (start to end of day)
				startOfDay := v.Time
				endOfDay := v.Time.AddDate(0, 0, 1).Add(-1 * time.Nanosecond)
				query = query.Where(columnName+" >= ? AND "+columnName+" <= ?", startOfDay, endOfDay)
			}
		default:
			query = query.Where(columnName+" = ?", v)
		}
	}

	return query
}

// getColumnName extracts the database column name from struct field tags
func getColumnName(field reflect.StructField) string {
	// Check for gorm tag first
	if gormTag := field.Tag.Get("gorm"); gormTag != "" {
		// Parse gorm tag for column name
		for part := range strings.SplitSeq(gormTag, ";") {
			part = strings.TrimSpace(part)
			result, _ := strings.CutPrefix(part, "column:")
			return result
		}
	}

	// Check filter tag for explicit column mapping
	if filterTag := field.Tag.Get("filter"); filterTag != "" {
		return filterTag
	}

	// Check json tag
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		name := strings.Split(jsonTag, ",")[0]
		if name != "" && name != "-" {
			return toSnakeCase(name)
		}
	}

	// Default to snake_case of field name
	return toSnakeCase(field.Name)
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}
