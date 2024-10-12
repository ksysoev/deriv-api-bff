package core

import (
	"html/template"
	"reflect"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handler"
)

func TestRequestProcessor_Render(t *testing.T) {
	tmpl, err := template.New("test").Parse("Params: {{.Params}}, ReqID: {{.ReqID}}")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	rp := &RequestProcessor{
		tempate: tmpl,
	}

	data := handler.TemplateData{
		Params: map[string]any{"key1": "value1", "key2": "value2"},
		ReqID:  12345,
	}
	expected := "Params: map[key1:value1 key2:value2], ReqID: 12345"

	result, err := rp.Render(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(result) != expected {
		t.Fatalf("expected %s, got %s", expected, string(result))
	}
}

func TestRequestProcessor_ParseResp_Success(t *testing.T) {
	rp := &RequestProcessor{
		responseBody: "data",
	}

	jsonData := `{"data": {"key1": "value1", "key2": "value2"}}`
	expected := map[string]any{"key1": "value1", "key2": "value2"}

	result, err := rp.ParseResp([]byte(jsonData))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestRequestProcessor_ParseResp_Error(t *testing.T) {
	rp := &RequestProcessor{
		responseBody: "data",
	}

	jsonData := `{"error": "something went wrong"}`

	_, err := rp.ParseResp([]byte(jsonData))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRequestProcessor_ParseResp_UnexpectedFormat(t *testing.T) {
	rp := &RequestProcessor{
		responseBody: "data",
	}

	jsonData := `{"data": 123}`

	result, err := rp.ParseResp([]byte(jsonData))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := map[string]any{"value": 123.0} // go will parse string numerics to float64
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}
