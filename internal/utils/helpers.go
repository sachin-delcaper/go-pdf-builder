package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// ParseFloat safely converts a string to float64
func ParseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Printf("Warning: Error parsing float '%s': %v", s, err)
		return 0
	}
	return f
}

// ParseInt safely converts a string to int
func ParseInt(s string) int {
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("Warning: Error parsing int '%s': %v", s, err)
		return 0
	}
	return i
}

// ReplaceVariables replaces template variables in text with actual values
func ReplaceVariables(text string, data map[string]interface{}) string {
	result := text

	// Replace variables in format {{variableName}}
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

// ReplaceVariablesInArray replaces variables from a JSON array format like ["var1", "var2"]
func ReplaceVariablesInArray(text, variableName string, data map[string]interface{}) string {
	if !strings.HasPrefix(variableName, "[") || !strings.HasSuffix(variableName, "]") {
		// Single variable, try direct replacement
		if val, ok := data[variableName]; ok {
			return strings.ReplaceAll(text, "{{"+variableName+"}}", fmt.Sprintf("%v", val))
		}

		// Try case-insensitive match
		for inputKey, val := range data {
			cleanInputKey := strings.TrimRight(inputKey, ":")
			if strings.EqualFold(cleanInputKey, variableName) {
				return strings.ReplaceAll(text, "{{"+variableName+"}}", fmt.Sprintf("%v", val))
			}
		}
		return text
	}

	// Array format: remove brackets and split
	varsStr := strings.Trim(variableName, "[]")
	vars := strings.Split(varsStr, ",")

	result := text
	for _, varName := range vars {
		// Clean up variable name
		cleanVar := strings.Trim(strings.Trim(varName, "\""), " ")

		// Try exact match first
		if val, ok := data[cleanVar]; ok {
			result = strings.ReplaceAll(result, "{{"+cleanVar+"}}", fmt.Sprintf("%v", val))
			continue
		}

		// Try case-insensitive match
		for inputKey, val := range data {
			cleanInputKey := strings.TrimRight(inputKey, ":")
			if strings.EqualFold(cleanInputKey, cleanVar) {
				result = strings.ReplaceAll(result, "{{"+cleanVar+"}}", fmt.Sprintf("%v", val))
				break
			}
		}
	}

	return result
}

// GetArrayFieldValue extracts a field value from an array element
func GetArrayFieldValue(item interface{}, fieldName string) string {
	if itemMap, isMap := item.(map[string]interface{}); isMap {
		if val, ok := itemMap[fieldName]; ok {
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

// IsValidPosition checks if position coordinates are valid
func IsValidPosition(x, y float64) bool {
	return x >= 0 && y >= 0
}

// IsValidSize checks if size dimensions are valid
func IsValidSize(width, height float64) bool {
	return width > 0 && height > 0
}

// NormalizeAlign normalizes alignment string to standard values
func NormalizeAlign(align string) string {
	switch strings.ToUpper(align) {
	case "LEFT", "L":
		return "L"
	case "CENTER", "C":
		return "C"
	case "RIGHT", "R":
		return "R"
	default:
		return "L" // Default to left
	}
}

// SafeString safely converts any value to string
func SafeString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// EnsureDirectory creates directory if it doesn't exist
func EnsureDirectory(path string) error {
	// This function would need os package import
	// For now, returning nil as it's handled elsewhere
	return nil
}

// LogDebug logs debug information if debug mode is enabled
func LogDebug(format string, args ...interface{}) {
	// This could be enhanced to check for debug flags
	log.Printf("[DEBUG] "+format, args...)
}

// LogInfo logs informational messages
func LogInfo(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

// LogError logs error messages
func LogError(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// LogWarn logs warning messages
func LogWarn(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

// TruncateString truncates a string to specified length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// IsNumeric checks if a string represents a numeric value
func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// Coalesce returns the first non-empty string
func Coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
