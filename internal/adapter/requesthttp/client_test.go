package requesthttp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

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
	response, err := client.Do(context.Background(), request.Request{
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
	response, err := client.Do(context.Background(), request.Request{
		URL:     url,
		Method:  request.MethodPost,
		Headers: request.NewHeadersFrom(map[string][]string{"X-Test": {"ok"}}),
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
	if got := response.Headers.Values("X-Reply"); len(got) != 1 || got[0] != "ok" {
		t.Fatalf("headers = %v, want X-Reply ok", response.Headers)
	}
	if got := response.Body; got != "done" {
		t.Fatalf("response body = %q, want %q", got, "done")
	}
	if got := response.URL.String(); got != server.URL {
		t.Fatalf("response url = %q, want %q", got, server.URL)
	}
}

func TestClientDoSendsRepeatedHeaders(t *testing.T) {
	var gotHeaders []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = append([]string(nil), r.Header.Values("X-Test")...)
	}))
	defer server.Close()

	url, err := request.NewURL(server.URL)
	if err != nil {
		t.Fatalf("new url: %v", err)
	}

	client := NewClient()
	_, err = client.Do(context.Background(), request.Request{
		URL:     url,
		Method:  request.MethodGet,
		Headers: request.NewHeadersFrom(map[string][]string{"X-Test": {"one", "two"}}),
	})
	if err != nil {
		t.Fatalf("do request: %v", err)
	}

	want := []string{"one", "two"}
	if !reflect.DeepEqual(gotHeaders, want) {
		t.Fatalf("X-Test headers = %v, want %v", gotHeaders, want)
	}
}

func TestNewClientSetsDefaultTimeout(t *testing.T) {
	client := NewClient()
	if client.httpClient == nil {
		t.Fatal("http client is nil")
	}
	if client.httpClient.Timeout != defaultRequestTimeout {
		t.Fatalf("timeout = %s, want %s", client.httpClient.Timeout, defaultRequestTimeout)
	}
}

func TestClientDoWorksWithoutConstructor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	url, err := request.NewURL(server.URL)
	if err != nil {
		t.Fatalf("new url: %v", err)
	}

	client := Client{}
	response, err := client.Do(context.Background(), request.Request{URL: url, Method: request.MethodGet})
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", response.StatusCode, http.StatusNoContent)
	}
}

func TestClientDoHonorsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	url, err := request.NewURL(server.URL)
	if err != nil {
		t.Fatalf("new url: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClient()
	_, err = client.Do(ctx, request.Request{URL: url, Method: request.MethodGet})
	if err == nil {
		t.Fatal("do request error is nil, want cancellation error")
	}
}

func TestClientDoTimesOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	url, err := request.NewURL(server.URL)
	if err != nil {
		t.Fatalf("new url: %v", err)
	}

	client := Client{httpClient: &http.Client{Timeout: time.Millisecond}}
	_, err = client.Do(context.Background(), request.Request{URL: url, Method: request.MethodGet})
	if err == nil {
		t.Fatal("do request error is nil, want timeout error")
	}
}
