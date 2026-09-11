package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("request = %s %s, content-type %q", r.Method, r.URL.Path, r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	var response struct {
		OK bool `json:"ok"`
	}
	if _, err := PostJSON(server.URL, map[string]string{"message": "hello"}, &response); err != nil {
		t.Fatal(err)
	}
	if !response.OK {
		t.Fatal("response was not decoded")
	}
}

func TestPostJSONRejectsInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	if _, err := PostJSON(server.URL, nil, &struct{}{}); err == nil {
		t.Fatal("PostJSON accepted an invalid response content type")
	}
}
