package generate

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseMarkdown(t *testing.T) {
	t.Run("valid markdown with full frontmatter", func(t *testing.T) {
		input := []byte(`---
title: Test Title
description: Test Description
lang: fr
tags: [tag1, tag2]
---
# Hello World
This is content.`)

		meta, html, err := parseMarkdown(input)

		if err != nil {
			t.Errorf("parseMarkdown() error = %v, wantErr nil", err)
		}

		expectedMeta := PageMeta{
			Title: "Test Title",
			Desc:  "Test Description",
			Lang:  "fr",
			Tags:  []string{"tag1", "tag2"},
		}
		if !reflect.DeepEqual(meta, expectedMeta) {
			t.Errorf("parseMarkdown() meta = %+v, want %+v", meta, expectedMeta)
		}

		if len(html) == 0 {
			t.Errorf("parseMarkdown() html is empty, want non-empty")
		}
		// Goldmark adds id attributes to headings
		if !strings.Contains(string(html), `id="hello-world"`) || !strings.Contains(string(html), ">Hello World</h1>") {
			t.Errorf("parseMarkdown() html = %s, want content containing '<h1 id=\"hello-world\">Hello World</h1>'", string(html))
		}
	})

	t.Run("valid markdown with minimal frontmatter (only title)", func(t *testing.T) {
		input := []byte(`---
title: Minimal Title
---
## Subheading
Minimal content.`)

		meta, html, err := parseMarkdown(input)

		if err != nil {
			t.Errorf("parseMarkdown() error = %v, wantErr nil", err)
		}

		expectedMeta := PageMeta{
			Title: "Minimal Title",
			Desc:  "",
			Lang:  "en", // Default
			Tags:  nil,  // Or empty slice, depending on YAML decoder
		}
		if meta.Title != expectedMeta.Title {
			t.Errorf("parseMarkdown() meta.Title = %q, want %q", meta.Title, expectedMeta.Title)
		}
		if meta.Lang != expectedMeta.Lang {
			t.Errorf("parseMarkdown() meta.Lang = %q, want %q", meta.Lang, expectedMeta.Lang)
		}
		if meta.Desc != expectedMeta.Desc {
			t.Errorf("parseMarkdown() meta.Desc = %q, want %q", meta.Desc, expectedMeta.Desc)
		}
		// Allow either nil or empty slice for Tags when not specified
		if meta.Tags != nil && len(meta.Tags) != 0 {
			t.Errorf("parseMarkdown() meta.Tags = %+v, want nil or empty", meta.Tags)
		}


		if len(html) == 0 {
			t.Errorf("parseMarkdown() html is empty, want non-empty")
		}
		// Goldmark adds id attributes to headings
		if !strings.Contains(string(html), `id="subheading"`) || !strings.Contains(string(html), ">Subheading</h2>") {
			t.Errorf("parseMarkdown() html = %s, want content containing '<h2 id=\"subheading\">Subheading</h2>'", string(html))
		}
	})

	t.Run("markdown missing frontmatter", func(t *testing.T) {
		input := []byte(`# Just Content
No frontmatter here.`)

		meta, html, err := parseMarkdown(input)

		if err != nil {
			t.Errorf("parseMarkdown() error = %v, wantErr nil", err)
		}

		// Expect default values
		expectedMeta := PageMeta{
			Title: "",
			Desc:  "",
			Lang:  "en", // Default
			Tags:  nil,
		}
		if !reflect.DeepEqual(meta, expectedMeta) {
			t.Errorf("parseMarkdown() meta = %+v, want %+v", meta, expectedMeta)
		}

		if len(html) == 0 {
			t.Errorf("parseMarkdown() html is empty, want non-empty")
		}
		// Goldmark adds id attributes to headings
		if !strings.Contains(string(html), `id="just-content"`) || !strings.Contains(string(html), ">Just Content</h1>") {
			t.Errorf("parseMarkdown() html = %s, want content containing '<h1 id=\"just-content\">Just Content</h1>'", string(html))
		}
	})

	t.Run("malformed frontmatter", func(t *testing.T) {
		input := []byte(`---
title: Test Title
description: Test Description
tags: [tag1, tag2
---
# Hello World
This is content.`) // Invalid YAML: unclosed bracket in tags

		_, _, err := parseMarkdown(input)

		if err == nil {
			t.Errorf("parseMarkdown() error = nil, wantErr for malformed frontmatter")
		}
		// Check if the error message indicates a frontmatter decoding issue
		// The actual error comes from the YAML parser used by goldmark-frontmatter
		if !strings.Contains(err.Error(), "error decoding frontmatter") && !strings.Contains(err.Error(), "yaml:") {
			t.Errorf("parseMarkdown() error = %v, want error related to frontmatter decoding or yaml", err)
		}
	})

	t.Run("lang field default and explicit", func(t *testing.T) {
		tests := []struct {
			name         string
			input        string
			expectedLang string
		}{
			{
				name: "lang omitted",
				input: `---
title: Lang Test
---
Content`,
				expectedLang: "en",
			},
			{
				name: "lang specified as fr",
				input: `---
title: Lang Test
lang: fr
---
Content`,
				expectedLang: "fr",
			},
			{
				name: "lang specified as empty string",
				input: `---
title: Lang Test
lang: "" 
---
Content`,
				expectedLang: "en", // Should default if empty string is provided
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				meta, _, err := parseMarkdown([]byte(tt.input))
				if err != nil {
					t.Fatalf("parseMarkdown() error = %v, wantErr nil for this case", err)
				}
				if meta.Lang != tt.expectedLang {
					t.Errorf("parseMarkdown() meta.Lang = %q, want %q", meta.Lang, tt.expectedLang)
				}
			})
		}
	})
}
