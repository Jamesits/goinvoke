//go:build windows

package goinvoke

import (
	"errors"
	"github.com/hashicorp/go-multierror"
	"github.com/jamesits/goinvoke/utils"
	"golang.org/x/sys/windows"
	"reflect"
	"strconv"
)

var typeOfLazyProc = reflect.TypeOf((*windows.LazyProc)(nil))
var typeOfProc = reflect.TypeOf((*windows.Proc)(nil))

// Unmarshal loads the DLL into memory, then fills all struct fields with type *windows.LazyProc with exported functions.
func Unmarshal(path string, v any) error {
	var err error
	var syntheticErr = errors.New("unmarshal failed")
	var errorOccurred = false

	var ld *windows.LazyDLL
	if utils.IsImplicitRelativePath(path) {
		ld = windows.NewLazySystemDLL(path)
	} else {
		ld = windows.NewLazyDLL(path)
	}

	err = ld.Load()
	if err != nil {
		errorOccurred = true
		syntheticErr = multierror.Append(syntheticErr, err)
		return syntheticErr
	}

	// create a corresponding windows.Dll object for compatibility
	d := &windows.DLL{
		Name:   ld.Name,
		Handle: windows.Handle(ld.Handle()),
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
		procName := utils.GetStructTag(typeField, "func")
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
			ordinal, ordinalParsingError := strconv.ParseInt(utils.GetStructTag(typeField, "ordinal"), 10, 64)

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
