//go:build windows

package goinvoke

import (
	"errors"
	"github.com/hashicorp/go-multierror"
	"github.com/jamesits/goinvoke/utils"
	"golang.org/x/sys/windows"
	"reflect"
	"strconv"
	"sync"
)

var typeOfLazyProc = reflect.TypeOf((*windows.LazyProc)(nil))
var typeOfProc = reflect.TypeOf((*windows.Proc)(nil))

var lazyDLLReferenceCache map[string]*windows.LazyDLL
var dllReferenceCache map[string]*windows.DLL
var globalDLLReferenceCacheWriteLock sync.Mutex

func init() {
	lazyDLLReferenceCache = map[string]*windows.LazyDLL{}
	dllReferenceCache = map[string]*windows.DLL{}
}

// Unmarshal loads the DLL into memory, then fills all struct fields with type *windows.LazyProc with exported functions.
func Unmarshal(path string, v any) error {
	var err error
	var syntheticErr = errors.New("unmarshal failed")
	var errorOccurred = false

	globalDLLReferenceCacheWriteLock.Lock()
	defer globalDLLReferenceCacheWriteLock.Unlock()

	// If multiple Unmarshal() is called with the same path string, the reference to the DLL will be cached.
	ld, ok := lazyDLLReferenceCache[path]
	if !ok {
		if utils.IsImplicitRelativePath(path) {
			ld = windows.NewLazySystemDLL(path)
		} else {
			ld = windows.NewLazyDLL(path)
		}
		lazyDLLReferenceCache[path] = ld
	}

	err = ld.Load()
	if err != nil {
		errorOccurred = true
		syntheticErr = multierror.Append(syntheticErr, err)
		return syntheticErr
	}

	// create a corresponding windows.Dll object for compatibility
	d, ok := dllReferenceCache[path]
	if !ok {
		d = &windows.DLL{
			Name:   ld.Name,
			Handle: windows.Handle(ld.Handle()),
		}
		dllReferenceCache[path] = d
	}

	// https://stackoverflow.com/a/46354875
	valueReference := reflect.ValueOf(v).Elem()
	typeReference := valueReference.Type()

	fieldCount := typeReference.NumField()
	for i := 0; i < fieldCount; i++ {
		// filter out incompatible attributes by their type
		t := valueReference.Field(i).Type()
		if (t != typeOfLazyProc) && (t != typeOfProc) {
			continue
		}

		// try to get a function name from tag first, then by attribute name
		typeField := typeReference.Field(i)
		procName := getStructTag(typeField, "func")
		if procName == "" {
			procName = typeField.Name
		}

		// get a reference of current attribute's value
		valueField := valueReference.Field(i)
		if !valueField.IsValid() || !valueField.CanSet() || procName == "" {
			continue
		}

		switch t {
		case typeOfLazyProc: // LazyProc only supports loading by name
			proc := ld.NewProc(procName)
			// try to load the proc now
			err = proc.Find()
			if err != nil {
				errorOccurred = true
				syntheticErr = multierror.Append(syntheticErr, err)
				continue
			}

			// https://stackoverflow.com/a/53110731
			valueField.Set(reflect.ValueOf(proc).Convert(valueField.Type()))

		case typeOfProc: // match by ordinal first, then name
			ordinal, ordinalParsingError := strconv.ParseInt(getStructTag(typeField, "ordinal"), 10, 64)

			var proc *windows.Proc
			if ordinalParsingError == nil { // we have a valid ordinal
				proc, err = d.FindProcByOrdinal(uintptr(ordinal))
			} else { // fallback to matching by name
				proc, err = d.FindProc(procName)
			}
			if err != nil {
				errorOccurred = true
				syntheticErr = multierror.Append(syntheticErr, err)
				continue
			}

			// https://stackoverflow.com/a/53110731
			valueField.Set(reflect.ValueOf(proc).Convert(valueField.Type()))
		default:
			continue
		}
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
