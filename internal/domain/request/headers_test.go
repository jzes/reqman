package request

import (
	"reflect"
	"testing"
)

func TestHeadersPreservesRepeatedValues(t *testing.T) {
	headers := NewHeaders()
	if err := headers.Add("Accept", "application/json"); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := headers.Add("Accept", "text/plain"); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	want := []string{"application/json", "text/plain"}
	if got := headers.Values("Accept"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Accept values = %v, want %v", got, want)
	}
}

func TestHeadersRowsAreSortedByName(t *testing.T) {
	headers := NewHeadersFrom(map[string][]string{
		"X-Zeta": {"last"},
		"Accept": {"application/json", "text/plain"},
	})

	want := []Header{
		{Key: "Accept", Value: "application/json"},
		{Key: "Accept", Value: "text/plain"},
		{Key: "X-Zeta", Value: "last"},
	}
	if got := headers.Rows(); !reflect.DeepEqual(got, want) {
		t.Fatalf("rows = %v, want %v", got, want)
	}
}

func TestHeadersAddIfMissingIsCaseInsensitive(t *testing.T) {
	headers := NewHeadersFrom(map[string][]string{"Content-Type": {"application/json"}})
	if err := headers.AddIfMissing("content-type", "text/plain"); err != nil {
		t.Fatalf("AddIfMissing() error = %v", err)
	}
	if err := headers.AddIfMissing("Accept", "application/json"); err != nil {
		t.Fatalf("AddIfMissing() error = %v", err)
	}

	if got := headers.Values("Content-Type"); !reflect.DeepEqual(got, []string{"application/json"}) {
		t.Fatalf("Content-Type values = %v, want [application/json]", got)
	}
	if got := headers.Values("Accept"); !reflect.DeepEqual(got, []string{"application/json"}) {
		t.Fatalf("Accept values = %v, want [application/json]", got)
	}
}

func TestHeadersValuesReturnsCopy(t *testing.T) {
	headers := NewHeadersFrom(map[string][]string{"Accept": {"application/json"}})
	values := headers.Values("Accept")
	values[0] = "text/plain"

	if got := headers.Values("Accept"); !reflect.DeepEqual(got, []string{"application/json"}) {
		t.Fatalf("Accept values = %v, want [application/json]", got)
	}
}

func TestHeadersAddRequiresInitialization(t *testing.T) {
	var headers Headers
	if err := headers.Add("Accept", "application/json"); err == nil {
		t.Fatal("Add() error = nil, want initialization error")
	}
}
