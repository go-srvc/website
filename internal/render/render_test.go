package render_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-srvc/website/internal/render"
)

func TestBuild(t *testing.T) {
	dir := t.TempDir()

	// Run from the website repo root so assets/static dirs are found.
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(filepath.Join(cwd, "..", "..")); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := render.Build(dir); err != nil {
		t.Fatalf("Build: %v", err)
	}

	idx, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	if !strings.Contains(string(idx), "go-srvc") {
		t.Errorf("index.html missing title text")
	}

	cname, err := os.ReadFile(filepath.Join(dir, "CNAME"))
	if err != nil {
		t.Fatalf("read CNAME: %v", err)
	}
	if strings.TrimSpace(string(cname)) != "go-srvc.com" {
		t.Errorf("unexpected CNAME content: %q", cname)
	}

	if _, err := os.Stat(filepath.Join(dir, "assets", "css", "main.css")); err != nil {
		t.Errorf("assets/css/main.css missing: %v", err)
	}
}
