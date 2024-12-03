//go:build unit

package utils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func fileExists(path string) bool {
	_, err := filepath.Abs(path)
	return err == nil
}

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -cover -run TestTemplate.*
func TestTemplate_FileVersion(t *testing.T) {
	SetAppDir()
	version := fileVersion("/css/output.css")
	if version == "" {
		t.Errorf("expected a valid version string, got empty")
	}
	version = fileVersion("/wrongFile")
	if version != "" {
		t.Errorf("expected a empty version string, got not empty")
	}
}

func TestAdd(t *testing.T) {
	if add(1.5, 2.5) != 4.0 {
		t.Error("add(1.5, 2.5) != 4.0")
	}
}

func TestAddInt(t *testing.T) {
	if addInt(1, 2) != 3 {
		t.Error("addInt(1, 2) != 3")
	}
}

func TestSub(t *testing.T) {
	if sub(5.0, 3.0) != 2.0 {
		t.Error("sub(5.0, 3.0) != 2.0")
	}
}

func TestTemplate_Dict(t *testing.T) {
	testCases := []struct {
		name      string
		input     []interface{}
		expected  TplData
		wantPanic bool
	}{
		{
			name:      "Valid dict creation",
			input:     []interface{}{"key1", "value1", "key2", 42},
			expected:  TplData{"key1": "value1", "key2": 42},
			wantPanic: false,
		},
		{
			name:      "Odd number of arguments",
			input:     []interface{}{"key1", "value1", "key2"},
			expected:  nil,
			wantPanic: true,
		},
		{
			name:      "Non-string key",
			input:     []interface{}{42, "value1", "key2", "value2"},
			expected:  nil,
			wantPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic, but got none")
					}
				}()
			}

			result := dict(tc.input...)

			if !tc.wantPanic {
				for k, v := range tc.expected {
					if result[k] != v {
						t.Errorf("Expected %v for key %s, got %v", v, k, result[k])
					}
				}
			}
		})
	}
}

func TestTemplate_CreateTemplate_BrokenComponents(t *testing.T) {
	SetAppDir()
	originalComponents := componentsPath
	componentsPath = componentsPath + "BrokenComponents"
	defer func() { componentsPath = originalComponents }()
	w := httptest.NewRecorder()
	templates := createTemplate(w, []string{"index"})
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for missing components, got %d", resp.StatusCode)
	}
	if templates != nil {
		t.Errorf("Expected templates to be nil due to error, but got %v", templates)
	}
}

func TestTemplate_RenderTemplate_Success(t *testing.T) {
	SetAppDir()
	w := httptest.NewRecorder()
	data := TplData{"Title": "Title test", "Message": "Hello from existing error.html"}
	RenderTemplate(w, []string{"error"}, data)
	resp := w.Result()
	body := w.Body.String()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(body, "Title test") {
		t.Errorf("expected rendered template to contain 'Title test', got %s", body)
	}
	if !strings.Contains(body, "Hello from existing error.html") {
		t.Errorf("expected rendered template to contain 'Hello from existing error.html', got %s", body)
	}
	if !strings.Contains(body, "<!doctype html>") {
		t.Errorf("expected rendered template to contain '<!doctype html>', got %s", body)
	}
}

func TestTemplate_RenderTemplate_WrongTpl(t *testing.T) {
	SetAppDir()
	w := httptest.NewRecorder()
	RenderTemplate(w, []string{"wrongTpl"}, TplData{})
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for invalid template, got %d", resp.StatusCode)
	}
}

func TestTemplate_RenderTemplate_BrokenLayout(t *testing.T) {
	SetAppDir()
	originalLayoutPath := layoutPath
	layoutPath = layoutPath + "BrokenLayoutPath"
	defer func() { layoutPath = originalLayoutPath }()
	w := httptest.NewRecorder()
	RenderTemplate(w, []string{"index"}, TplData{})
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for invalid template, got %d", resp.StatusCode)
	}
}

func TestTemplate_ExecuteTemplate_Error(t *testing.T) {
	SetAppDir()
	tmpl := template.New("test")
	_, err := tmpl.Parse(`{{ template "missingTemplate" . }}`)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	w := httptest.NewRecorder()
	executeTemplate(tmpl, w, "test", TplData{})
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for template execution error, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Error rendering template") {
		t.Errorf("Expected error message in response body, got: %s", body)
	}
}

func TestTemplate_RenderTemplateWithoutLayout_Success(t *testing.T) {
	SetAppDir()
	w := httptest.NewRecorder()
	data := TplData{"Title": "Title test", "Message": "Hello from existing error.html"}
	RenderTemplateWithoutLayout(w, []string{"error"}, "content", data)
	resp := w.Result()
	body := w.Body.String()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(body, "Title test") {
		t.Errorf("expected rendered template to contain 'Title test', got %s", body)
	}
	if !strings.Contains(body, "Hello from existing error.html") {
		t.Errorf("expected rendered template to contain 'Hello from existing error.html', got %s", body)
	}
	if strings.Contains(body, "<!doctype html>") {
		t.Errorf("expected rendered template to not contain '<!doctype html>', got %s", body)
	}
}

func TestTemplate_RenderBlockNeedLogin(t *testing.T) {
	SetAppDir()
	w := httptest.NewRecorder()
	RenderBlockNeedLogin(w)

	resp := w.Result()
	body := w.Body.String()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(body, "You need to be logged in") {
		t.Errorf("expected message about login, got %s", body)
	}
}
