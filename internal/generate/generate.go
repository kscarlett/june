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
	if err := d.Decode(&metadata); err != nil {
		return PageMeta{}, nil, fmt.Errorf("error decoding frontmatter: %w", err)
	}
	if metadata.Lang == "" {
		metadata.Lang = "en"
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

func Generate(cfg GenerateConfig) {
	source, err := os.ReadFile(cfg.Input)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(cfg.Output); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(cfg.Output), 0700)
	}

	metadata, generated, err := parseMarkdown(source)
	if err != nil {
		fmt.Println(err)
		return
	}

	if cfg.Ugc {
		p := bluemonday.UGCPolicy()
		generated = p.SanitizeBytes(generated)
	}

	tmplBytes := templatex.LoadTemplate(cfg.Template)
	styleBytes := templatex.LoadStyle(cfg.Style)
	css := string(styleBytes)

	tmpl, err := template.New("page").Parse(string(tmplBytes))
	if err != nil {
		panic(fmt.Sprintf("failed to parse template: %v", err))
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
		panic(fmt.Sprintf("failed to execute template: %v", err))
	}

	err = os.WriteFile(cfg.Output, out.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully wrote to %s\n", cfg.Output)
}
