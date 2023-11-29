package utils

import (
	"bufio"
	"os"
	"strings"
)

func PathsFromEnvironmentVariable(env string) []string {
	v := os.Getenv(env)
	if v == "" {
		return nil
	}

	return strings.Split(v, ":")
}

func PathsFromFileLines(file string) (ret []string) {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "include ") {
			ret = append(ret, PathsFromFileLines(strings.TrimLeft(line, "include "))...)
		} else {
			ret = append(ret, line)
		}
	}

	return
}
