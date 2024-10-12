package processor

import (
	"bytes"
	"html/template"
	"reflect"
	"testing"
)

func TestRequestProcessor_Render(t *testing.T) {
	tmpl, err := template.New("test").Parse("Params: {{.Params}}, ReqID: {{.ReqID}}")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	rp := &Processor{
		tmpl: tmpl,
	}

	params := map[string]any{"key1": "value1", "key2": "value2"}
	reqID := 12345
	expected := "Params: map[key1:value1 key2:value2], ReqID: 12345"

	var buf bytes.Buffer

	if err := rp.Render(&buf, int64(reqID), params); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if buf.String() != expected {
		t.Fatalf("expected %s, got %s", expected, buf.String())
	}
}

func TestRequestProcessor_parse_Success(t *testing.T) {
	rp := &Processor{
		responseBody: "data",
	}

	jsonData := `{"data": {"key1": "value1", "key2": "value2"}}`
	expected := map[string]any{"key1": "value1", "key2": "value2"}

	result, err := rp.parse([]byte(jsonData))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestRequestProcessor_parse_Error(t *testing.T) {
	rp := &Processor{
		responseBody: "data",
	}

	jsonData := `{"error": "something went wrong"}`

	_, err := rp.parse([]byte(jsonData))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRequestProcessor_parse_UnexpectedFormat(t *testing.T) {
	rp := &Processor{
		responseBody: "data",
	}

	jsonData := `{"data": 123}`

	result, err := rp.parse([]byte(jsonData))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := map[string]any{"value": 123.0} // go will parse string numerics to float64
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}
