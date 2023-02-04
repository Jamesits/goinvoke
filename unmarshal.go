//go:build windows

package goinvoke

import (
	"errors"
	"github.com/hashicorp/go-multierror"
	"github.com/jamesits/goinvoke/utils"
	"golang.org/x/sys/windows"
	"reflect"
	"sync"
)

var typeOfLazyProc = reflect.TypeOf((*windows.LazyProc)(nil))
var globalDllReferenceCache map[string]*windows.LazyDLL
var globalDLLReferenceCacheWriteLock sync.Mutex

func init() {
	globalDllReferenceCache = map[string]*windows.LazyDLL{}
}

// Unmarshal loads the DLL into memory, then fills all struct fields with type *windows.LazyProc with exported functions.
func Unmarshal(path string, v any) error {
	var err error
	var syntheticErr = errors.New("unmarshal failed")
	var errorOccurred = false

	globalDLLReferenceCacheWriteLock.Lock()
	defer globalDLLReferenceCacheWriteLock.Unlock()

	// If multiple Unmarshal() is called with the same path string, the reference to the DLL will be cached.
	d, ok := globalDllReferenceCache[path]
	if !ok {
		if utils.IsImplicitRelativePath(path) {
			d = windows.NewLazySystemDLL(path)
		} else {
			d = windows.NewLazyDLL(path)
		}
		globalDllReferenceCache[path] = d
	}

	err = d.Load()
	if err != nil {
		errorOccurred = true
		syntheticErr = multierror.Append(syntheticErr, err)
		return syntheticErr
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

		proc := d.NewProc(tag)
		// try to load the proc now
		err = proc.Find()
		if err != nil {
			errorOccurred = true
			syntheticErr = multierror.Append(syntheticErr, err)
			continue
		}

		// https://stackoverflow.com/a/53110731
		valueField.Set(reflect.ValueOf(proc).Convert(valueField.Type()))
	}

	if errorOccurred {
		return syntheticErr
	}

	return nil
}

// getStructTag returns the value of a named tag of a struct member
func getStructTag(f reflect.StructField, tagName string) string {
	return f.Tag.Get(tagName)
}
