package requestfile

import (
	"reflect"
	"testing"

	"github.com/jzes/reqman/internal/domain/request"
)

func parseCurlRequest(name, path string, content []byte) (request.Request, error) {
	return ParseCurlRequest(Item{Name: name, Path: path, Content: content})
}

func TestParseCurlRequestWithLocationAndURL(t *testing.T) {
	parsed, err := parseCurlRequest("req1.curl", "req1.curl", []byte("curl --location 'localhost:3000/books'"))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}

	if parsed.Name != "req1.curl" {
		t.Fatalf("Name = %q, want %q", parsed.Name, "req1.curl")
	}
	if parsed.Path != "req1.curl" {
		t.Fatalf("Path = %q, want %q", parsed.Path, "req1.curl")
	}
	if parsed.URL.String() != "localhost:3000/books" {
		t.Fatalf("URL = %q, want %q", parsed.URL.String(), "localhost:3000/books")
	}
	if parsed.Method != request.MethodGet {
		t.Fatalf("Method = %q, want %q", parsed.Method, request.MethodGet)
	}
	if parsed.Headers.Len() != 0 {
		t.Fatalf("Headers = %v, want empty", parsed.Headers)
	}
	if parsed.Body != "" {
		t.Fatalf("Body = %q, want empty", parsed.Body)
	}
}

func TestParseCurlRequestWithMethodHeaderAndJSONBody(t *testing.T) {
	parsed, err := parseCurlRequest("create.curl", "create.curl", []byte(`curl -X POST -H 'Content-Type: application/json' -d '{"name":"x"}' http://localhost/books`))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}

	if parsed.Method != request.MethodPost {
		t.Fatalf("Method = %q, want %q", parsed.Method, request.MethodPost)
	}
	assertHeaderValues(t, parsed.Headers, "Content-Type", []string{"application/json"})
	if parsed.Body != `{"name":"x"}` {
		t.Fatalf("Body = %q, want %q", parsed.Body, `{"name":"x"}`)
	}
	if parsed.URL.String() != "http://localhost/books" {
		t.Fatalf("URL = %q, want %q", parsed.URL.String(), "http://localhost/books")
	}
}

func TestParseCurlRequestInfersPostWhenJSONBodyExists(t *testing.T) {
	parsed, err := parseCurlRequest("create.curl", "create.curl", []byte(`curl -d '{"name":"x"}' http://localhost/books`))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}

	if parsed.Method != request.MethodPost {
		t.Fatalf("Method = %q, want %q", parsed.Method, request.MethodPost)
	}
}

func TestParseCurlRequestWithJSONFlagAddsContentType(t *testing.T) {
	parsed, err := parseCurlRequest("create.curl", "create.curl", []byte(`curl --json '{"name":"x"}' http://localhost/books`))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}

	if parsed.Method != request.MethodPost {
		t.Fatalf("Method = %q, want %q", parsed.Method, request.MethodPost)
	}
	assertHeaderValues(t, parsed.Headers, "Content-Type", []string{"application/json"})
	if parsed.Body != `{"name":"x"}` {
		t.Fatalf("Body = %q, want %q", parsed.Body, `{"name":"x"}`)
	}
}

func TestParseCurlRequestRejectsNonJSONBody(t *testing.T) {
	_, err := parseCurlRequest("create.curl", "create.curl", []byte(`curl -d 'name=x' http://localhost/books`))
	if err == nil {
		t.Fatal("parseCurlRequest() error = nil, want error")
	}
}

func TestParseCurlRequestRejectsMultipleBodies(t *testing.T) {
	_, err := parseCurlRequest("create.curl", "create.curl", []byte(`curl -d '{"a":1}' -d '{"b":2}' http://localhost/books`))
	if err == nil {
		t.Fatal("parseCurlRequest() error = nil, want error")
	}
}

func TestParseCurlRequestSupportsCompactAndEqualFlags(t *testing.T) {
	parsed, err := parseCurlRequest("create.curl", "create.curl", []byte(`curl -XPOST --header='X-Test: yes' --json='{"name":"x"}' http://localhost/books`))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}

	if parsed.Method != request.MethodPost {
		t.Fatalf("Method = %q, want %q", parsed.Method, request.MethodPost)
	}
	assertHeaderValues(t, parsed.Headers, "X-Test", []string{"yes"})
	assertHeaderValues(t, parsed.Headers, "Content-Type", []string{"application/json"})
}

func TestParseCurlRequestPreservesRepeatedHeaders(t *testing.T) {
	parsed, err := parseCurlRequest("req.curl", "req.curl", []byte(`curl -H 'Accept: application/json' -H 'Accept: text/plain' http://localhost/books`))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}

	assertHeaderValues(t, parsed.Headers, "Accept", []string{"application/json", "text/plain"})
}

func TestParseCurlRequestSkipsUnknownFlagWithValue(t *testing.T) {
	parsed, err := parseCurlRequest("req.curl", "req.curl", []byte(`curl --connect-timeout 10 http://localhost/books`))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}

	if parsed.URL.String() != "http://localhost/books" {
		t.Fatalf("URL = %q, want %q", parsed.URL.String(), "http://localhost/books")
	}
}

func TestParseCurlRequestSkipsUnknownShortBooleanFlag(t *testing.T) {
	parsed, err := parseCurlRequest("req.curl", "req.curl", []byte(`curl -s http://localhost/books`))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}

	if parsed.URL.String() != "http://localhost/books" {
		t.Fatalf("URL = %q, want %q", parsed.URL.String(), "http://localhost/books")
	}
}

func TestParseCurlRequestSupportsSavedJSONBodyWithEscapedNewlines(t *testing.T) {
	parsed, err := parseCurlRequest("create.curl", "create.curl", []byte(`curl -d "{\"name\":\"x\",\n\"active\":true}" http://localhost/books`))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}

	want := "{\"name\":\"x\",\n\"active\":true}"
	if parsed.Body != want {
		t.Fatalf("Body = %q, want %q", parsed.Body, want)
	}
}

func TestParseCurlRequestPreservesEscapedNewlinesInsideJSONString(t *testing.T) {
	parsed, err := parseCurlRequest("create.curl", "create.curl", []byte(`curl -d '{"message":"line 1\nline 2"}' http://localhost/books`))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}

	want := `{"message":"line 1\nline 2"}`
	if parsed.Body != want {
		t.Fatalf("Body = %q, want %q", parsed.Body, want)
	}
}

func TestParseCurlRequestAllowsEmptyURL(t *testing.T) {
	parsed, err := parseCurlRequest("req.curl", "req.curl", []byte(`curl -H 'Accept: application/json'`))
	if err != nil {
		t.Fatalf("parseCurlRequest() error = %v", err)
	}
	if got := parsed.URL.String(); got != "" {
		t.Fatalf("URL = %q, want empty", got)
	}
	assertHeaderValues(t, parsed.Headers, "Accept", []string{"application/json"})
}

func assertHeaderValues(t *testing.T, headers request.Headers, name string, want []string) {
	t.Helper()
	if got := headers.Values(name); !reflect.DeepEqual(got, want) {
		t.Fatalf("%s header = %v, want %v", name, got, want)
	}
}

func TestParseCurlRequestRequiresCurlCommand(t *testing.T) {
	_, err := parseCurlRequest("req.curl", "req.curl", []byte(`wget http://localhost/books`))
	if err == nil {
		t.Fatal("parseCurlRequest() error = nil, want error")
	}
}
