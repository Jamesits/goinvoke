package utils

import (
	"path/filepath"
	"strings"
)

// BaseName returns the base name portion of a path.
func BaseName(path string) string {
	// https://freshman.tech/snippets/go/filename-no-extension/
	path = filepath.Base(path)
	return path[:len(path)-len(filepath.Ext(path))]
}

// IsImplicitRelativePath tests if the path only contains a file name.
func IsImplicitRelativePath(path string) bool {
	return !(filepath.IsAbs(path) || strings.ContainsRune(filepath.Clean(path), filepath.Separator))
}
