package httpclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetAndPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/get" && r.Method == http.MethodGet {
			_, _ = w.Write([]byte("get"))
			return
		}
		if r.URL.Path == "/post" && r.Method == http.MethodPost && r.Header.Get(HdrKeyContentType) == HdrValApplicationJson {
			body, _ := io.ReadAll(r.Body)
			_, _ = w.Write(body)
			return
		}
		http.Error(w, "unexpected request", http.StatusBadRequest)
	}))
	defer server.Close()

	getResponse, err := Get(server.URL + "/get")
	if err != nil {
		t.Fatal(err)
	}
	defer getResponse.Body.Close()
	getBody, _ := io.ReadAll(getResponse.Body)
	if string(getBody) != "get" {
		t.Fatalf("GET body = %q", getBody)
	}

	postResponse, err := Post(server.URL+"/post", HdrValApplicationJson, strings.NewReader(`{"ok":true}`))
	if err != nil {
		t.Fatal(err)
	}
	defer postResponse.Body.Close()
	postBody, _ := io.ReadAll(postResponse.Body)
	if string(postBody) != `{"ok":true}` {
		t.Fatalf("POST body = %q", postBody)
	}
}

func TestPostMultipartForm(t *testing.T) {
	dir := t.TempDir()
	fileName := filepath.Join(dir, "payload.txt")
	if err := os.WriteFile(fileName, []byte("file contents"), 0o600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get(HdrKeyContentType), "multipart/form-data;") {
			t.Errorf("content type = %q", r.Header.Get(HdrKeyContentType))
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}
		var foundFile, foundParam bool
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			data, _ := io.ReadAll(part)
			if part.FormName() == "file" && part.FileName() == "payload.txt" && string(data) == "file contents" {
				foundFile = true
			}
			if part.FormName() == "name" && string(data) == "pmsg" {
				foundParam = true
			}
		}
		if !foundFile || !foundParam {
			t.Errorf("multipart parts = file:%t param:%t", foundFile, foundParam)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	response, err := PostMultipartForm(server.URL, NewMultipartForm().AddFile("file", fileName).AddParam("name", "pmsg"))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
}
