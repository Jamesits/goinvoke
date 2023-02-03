package main

import (
	_ "embed"
	"text/template"
)

//go:embed src.tmpl
var srcTemplateString string
var srcTemplate = template.Must(template.New("").Parse(srcTemplateString))

type export struct {
	TypeName string
	Function string
}

type templateData struct {
	ImportPath      string
	SelfPackageName string
	CommandLine     []string

	SelfGenerate bool

	DllFileName            string
	DestinationPackageName string
	TypeName               string

	Exports []export
}
