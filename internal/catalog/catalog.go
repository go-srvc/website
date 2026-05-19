// Package catalog defines the set of go-srvc packages exposed on the website
// and resolves their tagged versions.
package catalog

import (
	"fmt"
	"path/filepath"

	"github.com/go-srvc/website/internal/docparse"
	"github.com/go-srvc/website/internal/source"
)

// Pkg describes one Go package the site documents.
type Pkg struct {
	Repo       source.Repo
	Subdir     string // package directory relative to repo root; "." for whole-repo packages
	TagPrefix  string // tag prefix used to filter `git ls-remote --tags`; empty for whole-repo
	ImportPath string
	Slug       string // URL slug under Group, e.g. "tickermod"
	Group      string // URL group, e.g. "srvc" or "mods"
}

// URLPath returns the package's latest-version page path, relative to site root,
// without a leading slash. Pages prefix it with their own RelPrefix to navigate up.
func (p Pkg) URLPath() string {
	if p.Group == p.Slug {
		return p.Group + "/"
	}
	return p.Group + "/" + p.Slug + "/"
}

// All is the canonical list of packages rendered into the website.
var All = []Pkg{
	{
		Repo:       Srvc,
		Subdir:     ".",
		ImportPath: "github.com/go-srvc/srvc",
		Slug:       "srvc",
		Group:      "srvc",
	},
	modPkg("httpmod"),
	modPkg("logmod"),
	modPkg("metermod"),
	modPkg("sigmod"),
	modPkg("sqlmod"),
	modPkg("sqlxmod"),
	modPkg("tickermod"),
	modPkg("tracemod"),
}

// Srvc is the srvc git repository.
var Srvc = source.Repo{Name: "srvc", URL: "https://github.com/go-srvc/srvc.git"}

// Mods is the mods git repository.
var Mods = source.Repo{Name: "mods", URL: "https://github.com/go-srvc/mods.git"}

func modPkg(name string) Pkg {
	return Pkg{
		Repo:       Mods,
		Subdir:     name,
		TagPrefix:  name + "/",
		ImportPath: "github.com/go-srvc/mods/" + name,
		Slug:       name,
		Group:      "mods",
	}
}

// Versions returns the package's tagged versions, highest semver first.
func (p Pkg) Versions() ([]string, error) {
	return source.ListTags(p.Repo, p.TagPrefix)
}

// Fetch checks out the package at version and parses its documentation.
func (p Pkg) Fetch(cacheDir, version string) (*docparse.Package, error) {
	ref := p.TagPrefix + version
	dest, err := source.Checkout(cacheDir, p.Repo, ref)
	if err != nil {
		return nil, fmt.Errorf("checkout %s@%s: %w", p.Slug, version, err)
	}
	return docparse.Parse(filepath.Join(dest, p.Subdir), p.ImportPath)
}
