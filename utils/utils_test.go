package utils

import (
	"bytes"
	"io"
	"testing"
)

func TestWriteItemIfAny(t *testing.T) {
	s := "value"
	i := 42
	b := true
	cases := []struct {
		name  string
		item  interface{}
		key   string
		want  string
	}{
		{"string", s, "Key", "Key=value\n"},
		{"string ptr", &s, "Key", "Key=value\n"},
		{"nil string ptr", (*string)(nil), "Key", ""},
		{"int", i, "Key", "Key=42\n"},
		{"int ptr", &i, "Key", "Key=42\n"},
		{"nil int ptr", (*int)(nil), "Key", ""},
		{"bool", b, "Key", "Key=true\n"},
		{"bool ptr", &b, "Key", "Key=true\n"},
		{"nil bool ptr", (*bool)(nil), "Key", ""},
		{"nil interface", nil, "Key", ""},
		{"unsupported", 1.5, "Key", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			var w io.Writer = &buf
			WriteItemIfAny(tc.item, tc.key, &w)
			if buf.String() != tc.want {
				t.Errorf("got %q, want %q", buf.String(), tc.want)
			}
		})
	}
}

func TestWriteItemsIfAny(t *testing.T) {
	var buf bytes.Buffer
	var w io.Writer = &buf
	WriteItemsIfAny(nil, "Key", &w)
	if buf.String() != "" {
		t.Errorf("nil slice should write nothing, got %q", buf.String())
	}
	WriteItemsIfAny([]string{"a", "b"}, "Key", &w)
	if buf.String() != "Key=a\nKey=b\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestCreateDummyStringSlice(t *testing.T) {
	s := CreateDummyStringSlice()
	if len(s) != 1 || s[0] == "" {
		t.Errorf("unexpected dummy slice: %v", s)
	}
}
