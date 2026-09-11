package utils

import (
	"io"
	"strings"
	"testing"
)

func TestJsonEncodeDoesNotEscapeHTML(t *testing.T) {
	reader, err := JsonEncode(map[string]string{"value": "<b>&"})
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "{\"value\":\"<b>&\"}\n"; got != want {
		t.Fatalf("JsonEncode() = %q, want %q", got, want)
	}
}

func TestJsonDecode(t *testing.T) {
	var got struct {
		Name string `json:"name"`
	}
	if err := JsonDecode(strings.NewReader(`{"name":"pmsg"}`), &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "pmsg" {
		t.Fatalf("decoded name = %q", got.Name)
	}
	if err := JsonDecode(strings.NewReader("{}"), nil); err != nil {
		t.Fatalf("JsonDecode with nil target returned error: %v", err)
	}
}
