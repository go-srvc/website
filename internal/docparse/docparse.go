// Package docparse extracts package documentation from a Go source directory.
package docparse

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Package is the extracted documentation for one Go package.
type Package struct {
	ImportPath string
	Name       string
	Doc        string
	Consts     []ValueDecl
	Vars       []ValueDecl
	Funcs      []FuncDecl
	Types      []TypeDecl
	Examples   []Example
	Readme     string // raw markdown if README.md is present in dir
}

type ValueDecl struct {
	Names     []string
	Signature string
	Doc       string
}

type FuncDecl struct {
	Name      string
	Recv      string
	Signature string
	Doc       string
}

type TypeDecl struct {
	Name      string
	Signature string
	Doc       string
	Methods   []FuncDecl
}

type Example struct {
	Name string
	Doc  string
	Code string
}

// Parse reads the package at dir and returns its documentation.
// importPath is the canonical import path used as a hint for the doc package.
func Parse(dir, importPath string) (*Package, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, isGoFile, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", dir, err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no Go packages in %s", dir)
	}

	// Pick the primary package (the one whose name doesn't end in _test).
	var primary *ast.Package
	for name, p := range pkgs {
		if !strings.HasSuffix(name, "_test") {
			primary = p
			break
		}
	}
	if primary == nil {
		return nil, fmt.Errorf("no non-test package in %s", dir)
	}

	var nonTest, test []*ast.File
	for _, p := range pkgs {
		for fname, f := range p.Files {
			if strings.HasSuffix(fname, "_test.go") {
				test = append(test, f)
			} else {
				nonTest = append(nonTest, f)
			}
		}
	}

	dp, err := doc.NewFromFiles(fset, nonTest, importPath)
	if err != nil {
		return nil, fmt.Errorf("doc.NewFromFiles: %w", err)
	}
	exs := doc.Examples(test...)

	out := &Package{
		ImportPath: importPath,
		Name:       dp.Name,
		Doc:        dp.Doc,
	}

	for _, v := range dp.Consts {
		out.Consts = append(out.Consts, valueDecl(fset, v))
	}
	for _, v := range dp.Vars {
		out.Vars = append(out.Vars, valueDecl(fset, v))
	}
	for _, f := range dp.Funcs {
		out.Funcs = append(out.Funcs, funcDecl(fset, f))
	}
	for _, t := range dp.Types {
		td := TypeDecl{
			Name:      t.Name,
			Signature: typeSpecText(fset, t),
			Doc:       t.Doc,
		}
		for _, m := range t.Methods {
			td.Methods = append(td.Methods, funcDecl(fset, m))
		}
		out.Types = append(out.Types, td)
	}
	for _, ex := range exs {
		out.Examples = append(out.Examples, exampleFromDoc(fset, ex))
	}

	sort.Slice(out.Funcs, func(i, j int) bool { return out.Funcs[i].Name < out.Funcs[j].Name })
	sort.Slice(out.Types, func(i, j int) bool { return out.Types[i].Name < out.Types[j].Name })

	if md, err := os.ReadFile(filepath.Join(dir, "README.md")); err == nil {
		out.Readme = string(md)
	}

	return out, nil
}

func isGoFile(fi os.FileInfo) bool {
	return strings.HasSuffix(fi.Name(), ".go")
}

func valueDecl(fset *token.FileSet, v *doc.Value) ValueDecl {
	return ValueDecl{
		Names:     v.Names,
		Signature: nodeText(fset, v.Decl),
		Doc:       v.Doc,
	}
}

func funcDecl(fset *token.FileSet, f *doc.Func) FuncDecl {
	d := *f.Decl
	d.Body = nil
	recv := ""
	if d.Recv != nil && len(d.Recv.List) > 0 {
		recv = nodeText(fset, d.Recv.List[0].Type)
	}
	return FuncDecl{
		Name:      f.Name,
		Recv:      recv,
		Signature: nodeText(fset, &d),
		Doc:       f.Doc,
	}
}

func typeSpecText(fset *token.FileSet, t *doc.Type) string {
	if t.Decl == nil {
		return ""
	}
	// t.Decl is a *ast.GenDecl wrapping a TypeSpec. Print it whole.
	return nodeText(fset, t.Decl)
}

func nodeText(fset *token.FileSet, node ast.Node) string {
	var buf strings.Builder
	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 4}
	_ = cfg.Fprint(&buf, fset, node)
	return strings.TrimSpace(buf.String())
}

func exampleFromDoc(fset *token.FileSet, ex *doc.Example) Example {
	code := ""
	if ex.Play != nil {
		code = nodeText(fset, ex.Play)
	} else if ex.Code != nil {
		code = nodeText(fset, ex.Code)
	}
	name := "Example"
	if ex.Name != "" {
		name = "Example" + ex.Name
	}
	if ex.Suffix != "" {
		name += "_" + ex.Suffix
	}
	return Example{
		Name: name,
		Doc:  ex.Doc,
		Code: code,
	}
}
