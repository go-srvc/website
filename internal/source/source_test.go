package source_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-srvc/website/internal/source"
)

func TestListTags_srvc(t *testing.T) {
	if testing.Short() {
		t.Skip("network test")
	}
	tags, err := source.ListTags(source.Repo{
		Name: "srvc",
		URL:  "https://github.com/go-srvc/srvc.git",
	}, "")
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) == 0 {
		t.Fatal("expected at least one tag")
	}
	t.Logf("latest srvc: %s, total: %d", tags[0], len(tags))
}

func TestListTags_modsWithPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("network test")
	}
	tags, err := source.ListTags(source.Repo{
		Name: "mods",
		URL:  "https://github.com/go-srvc/mods.git",
	}, "tickermod/")
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) == 0 {
		t.Fatal("expected at least one tickermod tag")
	}
	for _, v := range tags {
		if v[0] != 'v' {
			t.Errorf("tag %q lacks v prefix", v)
		}
	}
}

func TestCheckout(t *testing.T) {
	if testing.Short() {
		t.Skip("network test")
	}
	cache := t.TempDir()
	dest, err := source.Checkout(cache, source.Repo{
		Name: "srvc",
		URL:  "https://github.com/go-srvc/srvc.git",
	}, "v0.2.0")
	if err != nil {
		t.Fatalf("Checkout: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "go.mod")); err != nil {
		t.Errorf("go.mod missing in checkout: %v", err)
	}
}
