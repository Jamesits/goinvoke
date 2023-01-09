package goinvoke

import (
	"golang.org/x/sys/windows"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

var typeOfLazyProc = reflect.TypeOf((*windows.LazyProc)(nil))
var globalDllReferenceCache map[string]*windows.LazyDLL
var globalDLLReferenceCacheWriteLock sync.Mutex

func init() {
	globalDllReferenceCache = map[string]*windows.LazyDLL{}
}

// Unmarshal loads the DLL into memory, then fills all struct fields with type *windows.LazyProc with exported functions.
// Due to security concerns, if the path is relative and only contains a base name (e.g. "kernel32.dll"), file lookup
// is limited to Windows system directories. If you want to load a DLL from your working directory, specify it
// explicitly with ".\\filename.dll".
func Unmarshal(path string, v any) (err error) {
	globalDLLReferenceCacheWriteLock.Lock()
	defer globalDLLReferenceCacheWriteLock.Unlock()

	d, ok := globalDllReferenceCache[path]
	if !ok {
		if filepath.IsAbs(path) || strings.ContainsRune(filepath.Clean(path), filepath.Separator) {
			d = windows.NewLazyDLL(path)
		} else {
			d = windows.NewLazySystemDLL(path)
		}
		globalDllReferenceCache[path] = d
	}

	// https://stackoverflow.com/a/46354875
	valueReference := reflect.ValueOf(v).Elem()
	typeReference := valueReference.Type()

	fieldCount := typeReference.NumField()
	for i := 0; i < fieldCount; i++ {
		if valueReference.Field(i).Type() != typeOfLazyProc {
			continue
		}

		typeField := typeReference.Field(i)
		tag := getStructTag(typeField, "func")
		if tag == "" {
			tag = typeField.Name
		}

		valueField := valueReference.Field(i)
		if !valueField.IsValid() || !valueField.CanSet() || tag == "" {
			continue
		}

		// https://stackoverflow.com/a/53110731
		valueField.Set(reflect.ValueOf(d.NewProc(tag)).Convert(valueField.Type()))
	}

	err = d.Load()
	return
}

// getStructTag returns the value of a named tag of a struct member
func getStructTag(f reflect.StructField, tagName string) string {
	return f.Tag.Get(tagName)
}
