// Package render builds the static site into an output directory.
package render

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed templates/*.html.tmpl
var templatesFS embed.FS

// Build wipes out and writes the full site under it.
func Build(out string) error {
	if err := os.RemoveAll(out); err != nil {
		return fmt.Errorf("clean output: %w", err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return fmt.Errorf("mkdir output: %w", err)
	}

	if err := renderIndex(out); err != nil {
		return fmt.Errorf("render index: %w", err)
	}
	if err := copyTree("assets", filepath.Join(out, "assets")); err != nil {
		return fmt.Errorf("copy assets: %w", err)
	}
	if err := copyTree("static", out); err != nil {
		return fmt.Errorf("copy static: %w", err)
	}
	return nil
}

type page struct {
	Title       string
	Description string
}

func renderIndex(out string) error {
	tmpl, err := template.ParseFS(templatesFS,
		"templates/layout.html.tmpl",
		"templates/index.html.tmpl",
	)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(out, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.ExecuteTemplate(f, "layout.html.tmpl", page{
		Title:       "go-srvc · Simple, Safe, Modular Service Runner",
		Description: "A tiny Go library for composing service modules with a clean lifecycle.",
	})
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
