package main

import (
	"errors"
	"fmt"
	"golang.org/x/tools/go/packages"
	"strings"
)

// getPackageName returns the package name of the files matching some patterns.
func getPackageName(patterns []string, tags []string) (string, error) {
	cfg := &packages.Config{
		Mode:       packages.NeedName,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return "", err
	}

	if len(pkgs) < 1 {
		return "", errors.New("no package found")
	}

	if len(pkgs) > 1 {
		return "", errors.New("multiple packages found")
	}

	return pkgs[0].Name, nil
}
