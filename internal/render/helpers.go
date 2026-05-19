package render

import (
	"bytes"
	"go/doc/comment"
	"html/template"

	"github.com/yuin/goldmark"
)

var funcs = template.FuncMap{
	"godoc":    godocHTML,
	"markdown": markdownHTML,
}

func godocHTML(src string) template.HTML {
	if src == "" {
		return ""
	}
	var p comment.Parser
	parsed := p.Parse(src)
	var pr comment.Printer
	return template.HTML(pr.HTML(parsed))
}

func markdownHTML(src string) template.HTML {
	if src == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := goldmark.New().Convert([]byte(src), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(src))
	}
	return template.HTML(buf.String())
}
