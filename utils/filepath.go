package utils

import (
	"os"
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
	return !(filepath.IsAbs(path) || strings.ContainsRune(path, filepath.Separator))
}

// ExecutableDir returns the directory containing current executable.
func ExecutableDir() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(p), nil
}
