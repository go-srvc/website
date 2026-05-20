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

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/go-srvc/website/internal/catalog"
	"github.com/go-srvc/website/internal/docparse"
)

//go:embed templates/*.html.tmpl
var templatesFS embed.FS

//go:embed example.go.txt
var practicalExample string

var (
	chromaStyle     = styles.Get("dracula")
	chromaFormatter = chromahtml.New(chromahtml.WithClasses(true), chromahtml.TabWidth(4))
)

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

	bundles, err := fetchAll(opts.Cache)
	if err != nil {
		return fmt.Errorf("fetch packages: %w", err)
	}

	if err := renderIndex(opts.Out, bundles); err != nil {
		return fmt.Errorf("render index: %w", err)
	}
	if len(bundles) > 0 {
		if err := renderModsCatalog(opts.Out, bundles); err != nil {
			return fmt.Errorf("render mods catalog: %w", err)
		}
		for _, b := range bundles {
			if err := renderPackage(opts.Out, b); err != nil {
				return fmt.Errorf("render %s: %w", b.Pkg.Slug, err)
			}
		}
	}

	if err := copyTree("assets", filepath.Join(opts.Out, "assets")); err != nil {
		return fmt.Errorf("copy assets: %w", err)
	}
	if err := copyTree("static", opts.Out); err != nil {
		return fmt.Errorf("copy static: %w", err)
	}
	if err := writeSyntaxCSS(opts.Out); err != nil {
		return fmt.Errorf("write syntax css: %w", err)
	}
	return nil
}

func highlightGo(src string) template.HTML {
	iter, err := lexers.Get("go").Tokenise(nil, src)
	if err != nil {
		return template.HTML(template.HTMLEscapeString(src))
	}
	var buf strings.Builder
	if err := chromaFormatter.Format(&buf, chromaStyle, iter); err != nil {
		return template.HTML(template.HTMLEscapeString(src))
	}
	return template.HTML(buf.String())
}

func writeSyntaxCSS(out string) error {
	dst := filepath.Join(out, "assets", "css", "syntax.css")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	return chromaFormatter.WriteCSS(f, chromaStyle)
}

// bundle groups a catalog entry with every tagged version's parsed docs.
type bundle struct {
	Pkg      catalog.Pkg
	Latest   string
	Versions []string // sorted highest semver first; same order rendered in dropdown
	Docs     map[string]*docparse.Package
}

func fetchAll(cache string) ([]bundle, error) {
	if cache == "" {
		return nil, nil
	}
	out := make([]bundle, 0, len(catalog.All))
	for _, p := range catalog.All {
		versions, err := p.Versions()
		if err != nil {
			return nil, fmt.Errorf("list %s tags: %w", p.Slug, err)
		}
		if len(versions) == 0 {
			slog.Warn("no tags found, skipping", slog.String("pkg", p.Slug))
			continue
		}
		b := bundle{
			Pkg:      p,
			Latest:   versions[0],
			Versions: versions,
			Docs:     make(map[string]*docparse.Package, len(versions)),
		}
		for _, v := range versions {
			slog.Info("fetching", slog.String("pkg", p.Slug), slog.String("version", v))
			doc, err := p.Fetch(cache, v)
			if err != nil {
				return nil, fmt.Errorf("fetch %s@%s: %w", p.Slug, v, err)
			}
			b.Docs[v] = doc
		}
		out = append(out, b)
	}
	return out, nil
}

type sidebarItem struct {
	Slug     string
	URL      string // relative to the page being rendered
	IsActive bool
}

type sidebarSection struct {
	Label    string
	URL      string
	IsActive bool
	Items    []sidebarItem
}

type indexData struct {
	Title       string
	Description string
	RelPrefix   string
	Sidebar     []sidebarSection
	SrvcVersion string
	ModCount    int
	ExampleHTML template.HTML
}

type modsCard struct {
	Pkg     catalog.Pkg
	Version string
	Doc     *docparse.Package
}

type modsData struct {
	Title       string
	Description string
	RelPrefix   string
	Sidebar     []sidebarSection
	Mods        []modsCard
}

type versionLink struct {
	Version  string
	URL      string // already includes the page's RelPrefix
	IsLatest bool
}

type pkgData struct {
	Title       string
	Description string
	RelPrefix   string
	Sidebar     []sidebarSection
	Pkg         catalog.Pkg
	Version     string
	IsLatest    bool
	Versions    []versionLink
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

func renderIndex(out string, bundles []bundle) error {
	data := indexData{
		Title:       "go-srvc · Simple, Safe, Modular Service Runner",
		Description: "A tiny Go library for composing service modules with a clean lifecycle.",
		RelPrefix:   "",
		Sidebar:     buildSidebar("", "", ""),
		ExampleHTML: highlightGo(practicalExample),
	}
	for _, b := range bundles {
		switch b.Pkg.Group {
		case "srvc":
			data.SrvcVersion = b.Latest
		case "mods":
			data.ModCount++
		}
	}
	tmpl := mustParse("templates/layout.html.tmpl", "templates/index.html.tmpl")
	return writeTemplate(tmpl, filepath.Join(out, "index.html"), data)
}

func renderModsCatalog(out string, bundles []bundle) error {
	var mods []modsCard
	for _, b := range bundles {
		if b.Pkg.Group != "mods" {
			continue
		}
		mods = append(mods, modsCard{
			Pkg:     b.Pkg,
			Version: b.Latest,
			Doc:     b.Docs[b.Latest],
		})
	}
	if len(mods) == 0 {
		return nil
	}
	tmpl := mustParse("templates/layout.html.tmpl", "templates/mods.html.tmpl")
	return writeTemplate(tmpl, filepath.Join(out, "mods", "index.html"), modsData{
		Title:       "Mods · go-srvc",
		Description: "Ready-made srvc modules: HTTP, SQL, OpenTelemetry, signal, ticker.",
		RelPrefix:   "../",
		Sidebar:     buildSidebar("../", "mods", ""),
		Mods:        mods,
	})
}

func renderPackage(out string, b bundle) error {
	tmpl := mustParse("templates/layout.html.tmpl", "templates/package.html.tmpl")
	base := filepath.Join(out, b.Pkg.Group)
	if b.Pkg.Slug != b.Pkg.Group {
		base = filepath.Join(base, b.Pkg.Slug)
	}

	for _, v := range b.Versions {
		doc := b.Docs[v]
		isLatest := v == b.Latest

		// Versioned page: dist/<group>[/slug]/<v>/index.html
		versionedPath := filepath.Join(base, v, "index.html")
		versionedPrefix := relPrefix(versionedPath, out)
		if err := writeTemplate(tmpl, versionedPath, makePkgData(b, v, doc, isLatest, versionedPrefix)); err != nil {
			return err
		}

		if isLatest {
			// Canonical "latest" page: dist/<group>[/slug]/index.html
			canonicalPath := filepath.Join(base, "index.html")
			canonicalPrefix := relPrefix(canonicalPath, out)
			if err := writeTemplate(tmpl, canonicalPath, makePkgData(b, v, doc, isLatest, canonicalPrefix)); err != nil {
				return err
			}
		}
	}
	return nil
}

func makePkgData(b bundle, v string, doc *docparse.Package, isLatest bool, prefix string) pkgData {
	versions := make([]versionLink, 0, len(b.Versions))
	for _, vv := range b.Versions {
		versions = append(versions, versionLink{
			Version:  vv,
			URL:      prefix + b.Pkg.URLPath() + vv + "/",
			IsLatest: vv == b.Latest,
		})
	}
	return pkgData{
		Title:       fmt.Sprintf("%s@%s · go-srvc", b.Pkg.ImportPath, v),
		Description: firstSentence(doc.Doc),
		RelPrefix:   prefix,
		Sidebar:     buildSidebar(prefix, b.Pkg.Group, b.Pkg.Slug),
		Pkg:         b.Pkg,
		Version:     v,
		IsLatest:    isLatest,
		Versions:    versions,
		Doc:         doc,
		DocHTML:     godocHTML(doc.Doc),
		ReadmeHTML:  markdownHTML(doc.Readme),
		Sections: pkgSections{
			HasConsts: len(doc.Consts) > 0,
			HasVars:   len(doc.Vars) > 0,
			HasFuncs:  len(doc.Funcs) > 0,
			HasTypes:  len(doc.Types) > 0,
			HasEx:     len(doc.Examples) > 0,
		},
	}
}

// relPrefix returns the relative path from a page's output file back to the
// site root, e.g. "../../" for dist/mods/tickermod/v1.0.0/index.html.
func relPrefix(outputPath, root string) string {
	rel, err := filepath.Rel(root, filepath.Dir(outputPath))
	if err != nil || rel == "." {
		return ""
	}
	depth := strings.Count(rel, string(filepath.Separator)) + 1
	return strings.Repeat("../", depth)
}

// buildSidebar returns the package tree shown in the layout. URLs are emitted
// relative to a page with the given prefix. The active section/item is the one
// matching activeGroup and (when set) activeSlug.
func buildSidebar(prefix, activeGroup, activeSlug string) []sidebarSection {
	type build struct {
		section *sidebarSection
		index   int
	}
	byGroup := map[string]*build{}
	var order []string

	for _, p := range catalog.All {
		b, ok := byGroup[p.Group]
		if !ok {
			b = &build{
				section: &sidebarSection{
					Label: p.Group,
					URL:   prefix + p.Group + "/",
				},
				index: len(order),
			}
			byGroup[p.Group] = b
			order = append(order, p.Group)
		}
		if p.Group == p.Slug {
			continue
		}
		b.section.Items = append(b.section.Items, sidebarItem{
			Slug:     p.Slug,
			URL:      prefix + p.URLPath(),
			IsActive: p.Group == activeGroup && p.Slug == activeSlug,
		})
	}

	out := make([]sidebarSection, len(order))
	for _, name := range order {
		sec := byGroup[name].section
		if name == activeGroup {
			sec.IsActive = true
		}
		out[byGroup[name].index] = *sec
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
