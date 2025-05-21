package generate

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"

	templatex "github.com/kscarlett/june/internal/template"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type PageMeta struct {
	Title string   `yaml:"title"`
	Desc  string   `yaml:"description"`
	Lang  string   `yaml:"lang"`
	Tags  []string `yaml:"tags"`
}

func VersionString() string {
	return fmt.Sprintf("june version %s - commit %s (built at %s)", version, commit, date)
}

func parseMarkdown(input []byte) (PageMeta, []byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM,
			extension.Typographer,
			extension.Footnote,
			&frontmatter.Extender{}),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	ctx := parser.NewContext()
	if err := md.Convert(input, &buf, parser.WithContext(ctx)); err != nil {
		return PageMeta{}, nil, err
	}

	var metadata PageMeta
	d := frontmatter.Get(ctx)

	if d == nil {
		// No frontmatter found, set defaults
		metadata.Lang = "en"
		// Other fields (Title, Desc, Tags) will be their zero values
	} else {
		// Frontmatter exists, try to decode it
		if err := d.Decode(&metadata); err != nil {
			return PageMeta{}, nil, fmt.Errorf("error decoding frontmatter: %w", err)
		}
		// Ensure lang defaults to "en" if specified as empty in frontmatter
		if metadata.Lang == "" {
			metadata.Lang = "en"
		}
	}

	return metadata, buf.Bytes(), nil
}

type GenerateConfig struct {
	Input    string
	Output   string
	Style    string
	Template string
	Ugc      bool
}

func Generate(cfg GenerateConfig) error {
	source, err := os.ReadFile(cfg.Input)
	if err != nil {
		return fmt.Errorf("failed to read input file %s: %w", cfg.Input, err)
	}

	// Ensure output directory exists
	outputDir := path.Dir(cfg.Output)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0755); err != nil { // Changed mode to 0755 for directories
			return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to stat output directory %s: %w", outputDir, err)
	}


	metadata, generated, err := parseMarkdown(source)
	if err != nil {
		return fmt.Errorf("failed to parse markdown: %w", err)
	}

	if cfg.Ugc {
		p := bluemonday.UGCPolicy()
		generated = p.SanitizeBytes(generated)
	}

	tmpl, err := templatex.LoadTemplate(cfg.Template)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	css, err := templatex.LoadStyle(cfg.Style)
	if err != nil {
		return fmt.Errorf("failed to load style: %w", err)
	}

	data := struct {
		PageMeta
		Content template.HTML
		Style   template.CSS
	}{
		PageMeta: metadata,
		Content:  template.HTML(generated),
		Style:    template.CSS(css),
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	err = os.WriteFile(cfg.Output, out.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file %s: %w", cfg.Output, err)
	}
	fmt.Printf("Successfully wrote to %s\n", cfg.Output)
	return nil
}
