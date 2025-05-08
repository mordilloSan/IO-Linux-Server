package templates

import (
	"embed"
	"html/template"
	"log"
)

//go:embed index.tmpl
var tmplFS embed.FS

var IndexTemplate *template.Template

func init() {
	var err error
	IndexTemplate, err = template.ParseFS(tmplFS, "index.tmpl")
	if err != nil {
		log.Fatalf("❌ Failed to parse embedded template: %v", err)
	}
}
