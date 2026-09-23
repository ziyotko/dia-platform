// Package storage owns the uploads directory: mapping public file URLs back to
// files on disk and cleaning them up when the owning DB row disappears.
package storage

import (
	"os"
	"path/filepath"
	"strings"

	"application/config"
)

// uploadRoot 是上传目录名（磁盘上的目录），外部访问前缀 = server.upload_dir_prefix + "/uploads"（见 main.go 的 r.Static）。
const uploadRoot = "uploads"

// LocalPath maps a stored file URL to a path on disk, reporting whether the URL is
// one of ours. Used by the authenticated download endpoints: uploads are no
// longer served as public static files, so every read is checked first.
func LocalPath(fileURL string) (string, bool) {
	return resolve(fileURL)
}

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

// resolve maps a public file URL to a path relative to the working directory.
// 兼容三种历史形态：`uploads/20240101/x.pdf`（相对）、`/uploads/20240101/x.pdf`（旧绝对）、
// `/business_application/uploads/...`（带当前部署前缀）。
// Anything outside the uploads directory — including traversal attempts such as
// "/uploads/../../config.yaml" — is rejected.
func resolve(fileURL string) (string, bool) {
	u := strings.TrimSpace(fileURL)
	if u == "" {
		return "", false
	}
	// Remote URLs belong to another host and are not ours to delete.
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return "", false
	}
	u = filepath.ToSlash(u)
	u = strings.TrimPrefix(u, "./")
	u = strings.TrimPrefix(u, "/")
	// 去掉部署前缀（如 business_application ），统一归一化到 uploads/ 开头
	if prefix := uploadPrefix(); prefix != "" {
		u = strings.TrimPrefix(u, prefix+string('/'))
	}
	clean := filepath.Clean(filepath.FromSlash(u))
	root := filepath.Clean(uploadRoot)
	if !strings.HasPrefix(clean, root+string(os.PathSeparator)) {
		return "", false
	}
	return clean, true
}

// uploadPrefix 返回 server.upload_dir_prefix 去掉首尾斜杠的形式（未配置时为空）。
func uploadPrefix() string {
	if config.Cfg == nil {
		return ""
	}
	return strings.Trim(strings.TrimSpace(config.Cfg.Server.UploadDirPrefix), "/")
}
