//go:build !windows

package utils

import "errors"

// GetSystemDirectory is not implemented on non-Windows OSes and always returns an error.
func GetSystemDirectory() (string, error) {
	return "", errors.New("searching DLL inside System32 is not implemented on current OS")
}
