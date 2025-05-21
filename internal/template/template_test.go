package templatex_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	templatex "github.com/kscarlett/june/internal/template"
)

func TestLoadTemplate(t *testing.T) {
	t.Run("loads user-provided template", func(t *testing.T) {
		tempDir := t.TempDir()
		validTemplateContent := `{{define "test"}}<h1>Hello {{.Name}}</h1>{{end}}`
		tempFile := filepath.Join(tempDir, "custom.gohtml")
		if err := os.WriteFile(tempFile, []byte(validTemplateContent), 0644); err != nil {
			t.Fatalf("Failed to create temp template file: %v", err)
		}

		tmpl, err := templatex.LoadTemplate(tempFile)
		if err != nil {
			t.Errorf("LoadTemplate() error = %v, wantErr nil", err)
		}
		if tmpl == nil {
			t.Fatalf("LoadTemplate() returned nil template, want non-nil")
		}
		// A simple check to see if a known definition from the template exists
		if tmpl.Lookup("test") == nil {
			t.Errorf("LoadTemplate() template does not contain expected definition 'test'")
		}
	})

	t.Run("falls back to embedded template if user path does not exist", func(t *testing.T) {
		tmpl, err := templatex.LoadTemplate("nonexistent/path/template.gohtml")
		if err != nil {
			t.Errorf("LoadTemplate() error = %v, wantErr nil (fallback expected)", err)
		}
		if tmpl == nil {
			t.Fatalf("LoadTemplate() returned nil template on fallback, want non-nil")
		}
		// Check if it's the default embedded template (e.g., by checking a known definition or name)
		// The default template is named "default" in the LoadTemplate function
		if tmpl.Name() != "default" {
			t.Errorf("LoadTemplate() expected fallback to template named 'default', got '%s'", tmpl.Name())
		}
	})

	t.Run("errors on invalid user-provided template", func(t *testing.T) {
		tempDir := t.TempDir()
		// More robustly invalid template content
		invalidTemplateContent := `{{define "test"}}<h1>Hello {{.Name}}</h1>{{end}}{{`
		tempFile := filepath.Join(tempDir, "invalid.gohtml")
		if err := os.WriteFile(tempFile, []byte(invalidTemplateContent), 0644); err != nil {
			t.Fatalf("Failed to create temp invalid template file: %v", err)
		}

		_, err := templatex.LoadTemplate(tempFile)
		if err == nil {
			t.Errorf("LoadTemplate() error = nil, wantErr for invalid template content")
		}
	})
}

func TestLoadStyle(t *testing.T) {
	t.Run("loads user-provided style", func(t *testing.T) {
		tempDir := t.TempDir()
		validStyleContent := `body { color: blue; }`
		tempFile := filepath.Join(tempDir, "custom.css")
		if err := os.WriteFile(tempFile, []byte(validStyleContent), 0644); err != nil {
			t.Fatalf("Failed to create temp style file: %v", err)
		}

		style, err := templatex.LoadStyle(tempFile)
		if err != nil {
			t.Errorf("LoadStyle() error = %v, wantErr nil", err)
		}
		if style != validStyleContent {
			t.Errorf("LoadStyle() style = %q, want %q", style, validStyleContent)
		}
	})

	t.Run("falls back to embedded style if user path does not exist", func(t *testing.T) {
		style, err := templatex.LoadStyle("nonexistent/path/style.css")
		if err != nil {
			t.Errorf("LoadStyle() error = %v, wantErr nil (fallback expected)", err)
		}
		if style == "" {
			t.Errorf("LoadStyle() returned empty style on fallback, want non-empty")
		}
		// A simple check for some known content from the embedded simple.css
		// For example, embeddedFiles.ReadFile("files/styles/simple.css")
		// Let's assume it contains "body {"
		if !strings.Contains(style, "body {") {
			t.Errorf("LoadStyle() fallback style does not seem to be the embedded one, content: %s", style)
		}
	})

	// Test for os.ReadFile error on an existing file is harder to reliably set up
	// without OS-level manipulations (e.g. changing permissions after stat).
	// The current LoadStyle first stats, then reads. If stat fails, it falls back.
	// If ReadFile fails after stat succeeded, it returns an error.
	// This case is implicitly covered if ReadFile fails for a valid path.
	// For example, if the file is deleted between Stat and ReadFile (race condition),
	// or if the file has permissions that allow Stat but not Read.
	// We can simulate the ReadFile error by making the file unreadable after stat,
	// but that's complex. A simpler approach for now is to ensure that if a file path *is* valid
	// and readable, it's loaded, and if it's not (and Stat fails), fallback occurs.
	// The "errors on invalid user-provided template" for LoadTemplate covers a similar
	// scenario where the file is readable but content is bad. For LoadStyle, content
	// validity isn't parsed like templates, just read.
}
