// Package render builds the static site into an output directory.
package render

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-srvc/website/internal/catalog"
	"github.com/go-srvc/website/internal/docparse"
)

//go:embed templates/*.html.tmpl
var templatesFS embed.FS

// Options configures a site build.
type Options struct {
	Out   string // output directory; wiped before build
	Cache string // cache directory for cloned source; empty disables doc rendering
}

// Build runs the full pipeline.
func Build(opts Options) error {
	if err := os.RemoveAll(opts.Out); err != nil {
		return fmt.Errorf("clean output: %w", err)
	}
	if err := os.MkdirAll(opts.Out, 0o755); err != nil {
		return fmt.Errorf("mkdir output: %w", err)
	}

	pkgs, err := fetchAll(opts.Cache)
	if err != nil {
		return fmt.Errorf("fetch packages: %w", err)
	}

	if err := renderIndex(opts.Out, pkgs); err != nil {
		return fmt.Errorf("render index: %w", err)
	}
	if len(pkgs) > 0 {
		if err := renderModsCatalog(opts.Out, pkgs); err != nil {
			return fmt.Errorf("render mods catalog: %w", err)
		}
		for _, p := range pkgs {
			if err := renderPackage(opts.Out, p); err != nil {
				return fmt.Errorf("render %s: %w", p.Pkg.Slug, err)
			}
		}
	}

	if err := copyTree("assets", filepath.Join(opts.Out, "assets")); err != nil {
		return fmt.Errorf("copy assets: %w", err)
	}
	if err := copyTree("static", opts.Out); err != nil {
		return fmt.Errorf("copy static: %w", err)
	}
	return nil
}

// rendered pairs a catalog entry with its resolved version and parsed docs.
type rendered struct {
	Pkg     catalog.Pkg
	Version string
	Doc     *docparse.Package
}

func fetchAll(cache string) ([]rendered, error) {
	if cache == "" {
		return nil, nil
	}
	out := make([]rendered, 0, len(catalog.All))
	for _, p := range catalog.All {
		versions, err := p.Versions()
		if err != nil {
			return nil, fmt.Errorf("list %s tags: %w", p.Slug, err)
		}
		if len(versions) == 0 {
			slog.Warn("no tags found, skipping", slog.String("pkg", p.Slug))
			continue
		}
		latest := versions[0]
		slog.Info("fetching package", slog.String("pkg", p.Slug), slog.String("version", latest))
		doc, err := p.Fetch(cache, latest)
		if err != nil {
			return nil, fmt.Errorf("fetch %s@%s: %w", p.Slug, latest, err)
		}
		out = append(out, rendered{Pkg: p, Version: latest, Doc: doc})
	}
	return out, nil
}

type indexData struct {
	Title       string
	Description string
	SrvcVersion string
	ModCount    int
}

type modsData struct {
	Title       string
	Description string
	Mods        []rendered
}

type pkgData struct {
	Title       string
	Description string
	Pkg         catalog.Pkg
	Version     string
	Doc         *docparse.Package
	DocHTML     template.HTML
	ReadmeHTML  template.HTML
	Sections    pkgSections
}

type pkgSections struct {
	HasConsts bool
	HasVars   bool
	HasFuncs  bool
	HasTypes  bool
	HasEx     bool
}

func renderIndex(out string, pkgs []rendered) error {
	data := indexData{
		Title:       "go-srvc · Simple, Safe, Modular Service Runner",
		Description: "A tiny Go library for composing service modules with a clean lifecycle.",
	}
	for _, p := range pkgs {
		switch p.Pkg.Group {
		case "srvc":
			data.SrvcVersion = p.Version
		case "mods":
			data.ModCount++
		}
	}
	tmpl := mustParse("templates/layout.html.tmpl", "templates/index.html.tmpl")
	return writeTemplate(tmpl, filepath.Join(out, "index.html"), data)
}

func renderModsCatalog(out string, pkgs []rendered) error {
	mods := filterGroup(pkgs, "mods")
	if len(mods) == 0 {
		return nil
	}
	tmpl := mustParse("templates/layout.html.tmpl", "templates/mods.html.tmpl")
	return writeTemplate(tmpl, filepath.Join(out, "mods", "index.html"), modsData{
		Title:       "Mods · go-srvc",
		Description: "Ready-made srvc modules: HTTP, SQL, OpenTelemetry, signal, ticker.",
		Mods:        mods,
	})
}

func renderPackage(out string, r rendered) error {
	tmpl := mustParse("templates/layout.html.tmpl", "templates/package.html.tmpl")
	path := filepath.Join(out, r.Pkg.Group)
	if r.Pkg.Slug != r.Pkg.Group {
		path = filepath.Join(path, r.Pkg.Slug)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	data := pkgData{
		Title:       fmt.Sprintf("%s · go-srvc", r.Pkg.ImportPath),
		Description: firstSentence(r.Doc.Doc),
		Pkg:         r.Pkg,
		Version:     r.Version,
		Doc:         r.Doc,
		DocHTML:     godocHTML(r.Doc.Doc),
		ReadmeHTML:  markdownHTML(r.Doc.Readme),
		Sections: pkgSections{
			HasConsts: len(r.Doc.Consts) > 0,
			HasVars:   len(r.Doc.Vars) > 0,
			HasFuncs:  len(r.Doc.Funcs) > 0,
			HasTypes:  len(r.Doc.Types) > 0,
			HasEx:     len(r.Doc.Examples) > 0,
		},
	}
	return writeTemplate(tmpl, filepath.Join(path, "index.html"), data)
}

func filterGroup(pkgs []rendered, group string) []rendered {
	var out []rendered
	for _, p := range pkgs {
		if p.Pkg.Group == group {
			out = append(out, p)
		}
	}
	return out
}

func mustParse(paths ...string) *template.Template {
	tmpl, err := template.New(filepath.Base(paths[0])).Funcs(funcs).ParseFS(templatesFS, paths...)
	if err != nil {
		panic(err)
	}
	return tmpl
}

func writeTemplate(tmpl *template.Template, path string, data any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.ExecuteTemplate(f, "layout.html.tmpl", data)
}

func copyTree(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.IndexAny(s, ".\n"); i > 0 {
		return strings.TrimSpace(s[:i+1])
	}
	return s
}
