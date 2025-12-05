package output

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

func TestJSONPrinter_PrintSuccess(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantJSON map[string]interface{}
	}{
		{
			name: "simple string data",
			data: map[string]string{"path": "/some/path"},
			wantJSON: map[string]interface{}{
				"success": true,
				"data":    map[string]interface{}{"path": "/some/path"},
			},
		},
		{
			name: "complex struct data",
			data: map[string]interface{}{
				"workspaces": []map[string]string{
					{"id": "ws1", "branch": "main"},
				},
			},
			wantJSON: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"workspaces": []interface{}{
						map[string]interface{}{"id": "ws1", "branch": "main"},
					},
				},
			},
		},
		{
			name: "nil data",
			data: nil,
			wantJSON: map[string]interface{}{
				"success": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			printer := NewJSONPrinter().WithWriter(&buf)

			err := printer.PrintSuccess(tt.data)
			if err != nil {
				t.Fatalf("PrintSuccess() error = %v", err)
			}

			var got map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("failed to unmarshal output: %v", err)
			}

			// Check success field
			if got["success"] != true {
				t.Errorf("success = %v, want true", got["success"])
			}

			// Check that error is not present or nil
			if got["error"] != nil {
				t.Errorf("error = %v, want nil", got["error"])
			}

			// Check that data matches expected payload
			wantData := tt.wantJSON["data"]

			gotData := got["data"]
			if !reflect.DeepEqual(gotData, wantData) {
				t.Errorf("data = %v, want %v", gotData, wantData)
			}
		})
	}
}

func TestJSONPrinter_PrintError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantCode     string
		wantContains string
	}{
		{
			name:         "canopy error",
			err:          cerrors.NewWorkspaceNotFound("test-ws"),
			wantCode:     "WORKSPACE_NOT_FOUND",
			wantContains: "test-ws",
		},
		{
			name:         "wrapped canopy error",
			err:          cerrors.WrapGitError(nil, "clone"),
			wantCode:     "GIT_OPERATION_FAILED",
			wantContains: "clone",
		},
		{
			name:         "generic error falls back to internal",
			err:          cerrors.NewInternalError("something went wrong", nil),
			wantCode:     "INTERNAL_ERROR",
			wantContains: "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			printer := NewJSONPrinter().WithWriter(&buf)

			err := printer.PrintError(tt.err)
			if err != nil {
				t.Fatalf("PrintError() error = %v", err)
			}

			var got Response
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("failed to unmarshal output: %v", err)
			}

			if got.Success {
				t.Error("success = true, want false")
			}

			if got.Error == nil {
				t.Fatal("error = nil, want non-nil")
			}

			if got.Error.Code != tt.wantCode {
				t.Errorf("error.code = %v, want %v", got.Error.Code, tt.wantCode)
			}

			if got.Error.Message == "" {
				t.Error("error.message is empty, want non-empty")
			}

			if !strings.Contains(got.Error.Message, tt.wantContains) {
				t.Errorf("error.message = %q, want to contain %q", got.Error.Message, tt.wantContains)
			}
		})
	}
}

func TestPrintJSON(t *testing.T) {
	// Verify that PrintJSON returns valid JSON envelope format
	var buf bytes.Buffer

	printer := NewJSONPrinter().WithWriter(&buf)

	data := map[string]string{"key": "value"}
	if err := printer.PrintSuccess(data); err != nil {
		t.Fatalf("PrintSuccess() error = %v", err)
	}

	var response Response
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if !response.Success {
		t.Error("success should be true")
	}
}

func TestJSONIndentation(t *testing.T) {
	var buf bytes.Buffer

	printer := NewJSONPrinter().WithWriter(&buf)

	data := map[string]string{"a": "b"}
	if err := printer.PrintSuccess(data); err != nil {
		t.Fatalf("PrintSuccess() error = %v", err)
	}

	// Generate expected output with 2-space indentation
	var expected bytes.Buffer

	enc := json.NewEncoder(&expected)
	enc.SetIndent("", "  ")
	_ = enc.Encode(Response{Success: true, Data: data})

	if buf.String() != expected.String() {
		t.Errorf("indentation mismatch:\ngot:\n%s\nwant:\n%s", buf.String(), expected.String())
	}
}

func TestErrorToInfo_NilError(t *testing.T) {
	result := errorToInfo(nil)
	if result != nil {
		t.Errorf("errorToInfo(nil) = %v, want nil", result)
	}
}
