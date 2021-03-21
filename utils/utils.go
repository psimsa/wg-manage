package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

// HandleError is a shortcut for logging an error as "fatal" with a custom message.
func HandleError(err error, message string) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func SaveToFile(name string, data []byte) {
	file, err := os.Create(name)
	HandleError(err, "Could not create file")
	defer file.Close()
	_, _ = file.Write(data)
}

func CreateDummyStringSlice() []string {
	slice := make([]string, 1)
	slice[0] = "<add here or remove section>"
	return slice
}

func WriteItemsIfAny(slice []string, key string, w *io.Writer) {
	if len(slice) > 0 {
		for _, v := range slice {
			fmt.Fprintf(*w, "%s=%s", key, v)
			fmt.Fprintln(*w)
		}
	}
}
func WriteItemIfAny(item interface{}, key string, w *io.Writer) {

	if item != nil {
		switch typed := item.(type) {
		case string:
			fmt.Fprintf(*w, "%s=%s", key, typed)
		case *string:
			if typed == nil {
				return
			}
			fmt.Fprintf(*w, "%s=%s", key, *typed)
		case int:
			fmt.Fprintf(*w, "%s=%s", key, strconv.Itoa(typed))
		case *int:
			if typed == nil {
				return
			}
			fmt.Fprintf(*w, "%s=%s", key, strconv.Itoa(*typed))
		case bool:
			fmt.Fprintf(*w, "%s=%s", key, strconv.FormatBool(typed))
		case *bool:
			if typed == nil {
				return
			}
			fmt.Fprintf(*w, "%s=%s", key, strconv.FormatBool(*typed))
		default:
			return
		}

		fmt.Fprintln(*w)
	}
}
