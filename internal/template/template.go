package templatex

import (
	"embed"
	"html/template"
	"os"
)

var (
	//go:embed files/*
	embeddedFiles embed.FS
)

func LoadTemplate(templatePath string) (*template.Template, error) {
	if _, err := os.Stat(templatePath); err == nil {
		b, err := os.ReadFile(templatePath)
		if err != nil {
			return nil, err
		}
		tmpl, err := template.New("custom").Parse(string(b))
		if err != nil {
			return nil, err
		}
		return tmpl, nil
	}
	b, err := embeddedFiles.ReadFile("files/templates/basic.gohtml")
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("default").Parse(string(b))
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func LoadStyle(stylePath string) (string, error) {
	if _, err := os.Stat(stylePath); err == nil {
		b, err := os.ReadFile(stylePath)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	b, err := embeddedFiles.ReadFile("files/styles/simple.css")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
