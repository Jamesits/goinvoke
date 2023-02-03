//go:build windows

package utils

import "golang.org/x/sys/windows"

// GetSystemDirectory returns the absolute path of "System32" directory.
func GetSystemDirectory() (string, error) {
	return windows.GetSystemDirectory()
}
