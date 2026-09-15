package jsoneditor

import "testing"

func TestFormatJSONFormatsValidJSON(t *testing.T) {
	panel := NewPanel()
	panel.SetValue(`{"a":[1,true]}`)

	if ok := panel.FormatJSON(); !ok {
		t.Fatal("FormatJSON() = false, want true")
	}

	want := "{\n  \"a\": [\n    1,\n    true\n  ]\n}"
	if got := panel.Value(); got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}

func TestFormatJSONInvalidJSONDoesNotChangeValue(t *testing.T) {
	panel := NewPanel()
	panel.SetValue(`{"bad":`)

	if ok := panel.FormatJSON(); ok {
		t.Fatal("FormatJSON() = true, want false")
	}

	want := `{"bad":`
	if got := panel.Value(); got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}

func TestPersistChangesFormatsJSON(t *testing.T) {
	panel := NewPanel()
	panel.SetValue(`{"a":[1,true]}`)

	panel.PersistChanges()

	want := "{\n  \"a\": [\n    1,\n    true\n  ]\n}"
	if got := panel.Value(); got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}

func TestPersistChangesInvalidJSONDoesNotChangeValue(t *testing.T) {
	panel := NewPanel()
	panel.SetValue(`{"bad":`)

	panel.PersistChanges()

	want := `{"bad":`
	if got := panel.Value(); got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}
