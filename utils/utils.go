package utils

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

func SaveToFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0o644)
}

func CreateDummyStringSlice() []string {
	return []string{"<add here or remove section>"}
}

func WriteItemsIfAny(slice []string, key string, w io.Writer) {
	for _, value := range slice {
		fmt.Fprintf(w, "%s=%s\n", key, value)
	}
}

func WriteItemIfAny(item any, key string, w io.Writer) {
	if item == nil {
		return
	}

	if value, ok := formatItem(item); ok {
		fmt.Fprintf(w, "%s=%s\n", key, value)
	}
}

func formatItem(item any) (string, bool) {
	switch typed := item.(type) {
	case string:
		return typed, true
	case *string:
		if typed == nil {
			return "", false
		}
		return *typed, true
	case int:
		return strconv.Itoa(typed), true
	case *int:
		if typed == nil {
			return "", false
		}
		return strconv.Itoa(*typed), true
	case bool:
		return strconv.FormatBool(typed), true
	case *bool:
		if typed == nil {
			return "", false
		}
		return strconv.FormatBool(*typed), true
	default:
		return "", false
	}
}
