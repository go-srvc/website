package docparse_test

import (
	"strings"
	"testing"

	"github.com/go-srvc/website/internal/docparse"
)

func TestParse_sample(t *testing.T) {
	pkg, err := docparse.Parse("testdata/sample", "github.com/go-srvc/website/internal/docparse/testdata/sample")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if pkg.Name != "sample" {
		t.Errorf("Name = %q, want sample", pkg.Name)
	}
	if !strings.Contains(pkg.Doc, "fixture") {
		t.Errorf("Doc missing expected content: %q", pkg.Doc)
	}

	if got := len(pkg.Consts); got != 1 {
		t.Errorf("Consts len = %d, want 1", got)
	}
	if got := len(pkg.Vars); got != 1 {
		t.Errorf("Vars len = %d, want 1", got)
	}

	wantFunc := map[string]bool{"Greet": false}
	for _, f := range pkg.Funcs {
		if _, ok := wantFunc[f.Name]; ok {
			wantFunc[f.Name] = true
		}
	}
	for name, found := range wantFunc {
		if !found {
			t.Errorf("missing exported func %s", name)
		}
	}

	var counter *docparse.TypeDecl
	for i := range pkg.Types {
		if pkg.Types[i].Name == "Counter" {
			counter = &pkg.Types[i]
		}
	}
	if counter == nil {
		t.Fatal("missing type Counter")
	}
	if got := len(counter.Methods); got != 2 {
		t.Errorf("Counter methods len = %d, want 2", got)
	}

	if got := len(pkg.Examples); got != 1 {
		t.Errorf("Examples len = %d, want 1", got)
	}

	if pkg.Readme == "" {
		t.Errorf("Readme empty; expected README.md content")
	}
}
