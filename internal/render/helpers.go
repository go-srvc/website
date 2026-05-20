package render

import (
	"bytes"
	"go/doc/comment"
	"html/template"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

var funcs = template.FuncMap{
	"godoc":    godocHTML,
	"markdown": markdownHTML,
	"hl":       highlightGo,
}

var markdown = goldmark.New(
	goldmark.WithExtensions(
		highlighting.NewHighlighting(
			highlighting.WithStyle("dracula"),
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
				chromahtml.TabWidth(4),
			),
		),
	),
)

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
	if err := markdown.Convert([]byte(src), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(src))
	}
	return template.HTML(buf.String())
}
