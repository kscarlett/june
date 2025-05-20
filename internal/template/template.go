package templatex

import (
	"embed"
	"os"
)

var (
	//go:embed files/*
	embeddedFiles embed.FS
)

func LoadTemplate(templatePath string) []byte {
	if _, err := os.Stat(templatePath); err == nil {
		b, err := os.ReadFile(templatePath)
		if err != nil {
			panic(err)
		}
		return b
	}
	b, err := embeddedFiles.ReadFile("files/templates/basic.gohtml")
	if err != nil {
		panic(err)
	}
	return b
}

func LoadStyle(stylePath string) []byte {
	if _, err := os.Stat(stylePath); err == nil {
		b, err := os.ReadFile(stylePath)
		if err != nil {
			panic(err)
		}
		return b
	}
	b, err := embeddedFiles.ReadFile("files/styles/simple.css")
	if err != nil {
		panic(err)
	}
	return b
}
