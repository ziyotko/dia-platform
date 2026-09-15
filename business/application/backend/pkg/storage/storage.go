// Package storage owns the uploads directory: mapping public file URLs back to
// files on disk and cleaning them up when the owning DB row disappears.
package storage

import (
	"os"
	"path/filepath"
	"strings"
)

// uploadRoot mirrors the static mount in main.go (`r.Static("/uploads", "./uploads")`).
const uploadRoot = "uploads"

// RemoveByURL deletes the file a stored file URL points at. Best effort: an
// already missing file (or a URL we do not manage) is not an error, so callers
// never have to branch on it.
func RemoveByURL(fileURL string) {
	p, ok := resolve(fileURL)
	if !ok {
		return
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		// Leftover files are harmless; a failed delete must never fail the
		// request that triggered it.
		return
	}
}

// RemoveAll is RemoveByURL for a batch of URLs.
func RemoveAll(fileURLs []string) {
	for _, u := range fileURLs {
		RemoveByURL(u)
	}
}

// resolve maps a public file URL ("/uploads/20240101/x.pdf") to a path relative
// to the working directory. Anything outside the uploads directory — including
// traversal attempts such as "/uploads/../../config.yaml" — is rejected.
func resolve(fileURL string) (string, bool) {
	u := strings.TrimSpace(fileURL)
	if u == "" {
		return "", false
	}
	// Remote URLs belong to another host and are not ours to delete.
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return "", false
	}
	clean := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(u, "/")))
	root := filepath.Clean(uploadRoot)
	if !strings.HasPrefix(clean, root+string(os.PathSeparator)) {
		return "", false
	}
	return clean, true
}
