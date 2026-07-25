package requesthttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jzes/reqman/internal/domain/request"
)

func TestClientDoAcceptsURLWithoutScheme(t *testing.T) {
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
	}))
	defer server.Close()

	serverURL := strings.TrimPrefix(server.URL, "http://")
	url, err := request.NewURL(serverURL + "/books")
	if err != nil {
		t.Fatalf("new url: %v", err)
	}

	client := NewClient()
	response, err := client.Do(request.Request{
		URL:    url,
		Method: request.MethodGet,
	})
	if err != nil {
		t.Fatalf("do request: %v", err)
	}

	if gotPath != "/books" {
		t.Fatalf("path = %q, want /books", gotPath)
	}
	if got := response.URL.String(); got != server.URL+"/books" {
		t.Fatalf("response url = %q, want %q", got, server.URL+"/books")
	}
}

func TestClientDo(t *testing.T) {
	var gotMethod string
	var gotHeader string
	var gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}

		gotMethod = r.Method
		gotHeader = r.Header.Get("X-Test")
		gotBody = string(body)
		w.Header().Set("X-Reply", "ok")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("done"))
	}))
	defer server.Close()

	url, err := request.NewURL(server.URL)
	if err != nil {
		t.Fatalf("new url: %v", err)
	}

	client := NewClient()
	response, err := client.Do(request.Request{
		URL:     url,
		Method:  request.MethodPost,
		Headers: map[string]string{"X-Test": "ok"},
		Body:    "hello",
	})
	if err != nil {
		t.Fatalf("do request: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotHeader != "ok" {
		t.Fatalf("header = %q, want %q", gotHeader, "ok")
	}
	if gotBody != "hello" {
		t.Fatalf("body = %q, want %q", gotBody, "hello")
	}
	if got := response.StatusCode; got != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", got, http.StatusCreated)
	}
	if got := response.Status; got != "201 Created" {
		t.Fatalf("status = %q, want %q", got, "201 Created")
	}
	if got := response.Headers["X-Reply"]; len(got) != 1 || got[0] != "ok" {
		t.Fatalf("headers = %v, want X-Reply ok", response.Headers)
	}
	if got := response.Body; got != "done" {
		t.Fatalf("response body = %q, want %q", got, "done")
	}
	if got := response.URL.String(); got != server.URL {
		t.Fatalf("response url = %q, want %q", got, server.URL)
	}
}
