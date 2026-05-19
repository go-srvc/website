// Package source fetches go-srvc repositories at specific tags into a cache directory.
package source

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
)

// Repo identifies a remote git repository.
type Repo struct {
	Name string // short identifier, used as cache subdirectory
	URL  string // git URL passed to `git clone`
}

// ListTags returns tags whose ref name starts with prefix, sorted highest semver first.
// The returned values keep the prefix stripped (e.g. "v1.0.1", not "tickermod/v1.0.1").
func ListTags(repo Repo, prefix string) ([]string, error) {
	out, err := exec.Command("git", "ls-remote", "--tags", "--refs", repo.URL).Output()
	if err != nil {
		return nil, fmt.Errorf("ls-remote %s: %w", repo.URL, err)
	}
	full := prefix
	var tags []string
	for line := range strings.Lines(string(out)) {
		_, ref, ok := strings.Cut(strings.TrimSpace(line), "refs/tags/")
		if !ok {
			continue
		}
		if !strings.HasPrefix(ref, full) {
			continue
		}
		v := strings.TrimPrefix(ref, full)
		if !semver.IsValid(v) {
			continue
		}
		tags = append(tags, v)
	}
	slices.SortFunc(tags, func(a, b string) int { return semver.Compare(b, a) })
	return tags, nil
}

// Checkout clones repo at tag (full ref name) into dest as a shallow checkout.
// If dest already exists it is returned untouched.
func Checkout(cacheDir string, repo Repo, ref string) (string, error) {
	dest := filepath.Join(cacheDir, repo.Name, sanitize(ref))
	if _, err := os.Stat(dest); err == nil {
		return dest, nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	cmd := exec.Command("git", "clone", "--depth=1", "--branch", ref, repo.URL, dest)
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(dest)
		return "", fmt.Errorf("clone %s@%s: %w: %s", repo.URL, ref, err, out)
	}
	return dest, nil
}

func sanitize(ref string) string {
	return strings.ReplaceAll(ref, "/", "_")
}
