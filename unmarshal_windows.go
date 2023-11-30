//go:build windows

package goinvoke

import (
	"github.com/hashicorp/go-multierror"
	"github.com/jamesits/goinvoke/utils"
	"golang.org/x/sys/windows"
	"reflect"
	"strconv"
)

// Unmarshal loads the DLL into memory, then fills all struct fields with type *windows.LazyProc with exported functions.
func Unmarshal(path string, v any) error {
	var err error
	var syntheticErr = ErrorUnmarshalFailed
	var errorOccurred = false

	ld := newLazyDLL(path)
	err = ld.Load()
	if err != nil {
		errorOccurred = true
		syntheticErr = multierror.Append(syntheticErr, err)
		return syntheticErr
	}
	d := unLazy(ld)

	// https://stackoverflow.com/a/46354875
	valueReference := reflect.ValueOf(v).Elem()
	typeReference := valueReference.Type()

	fieldCount := typeReference.NumField()
	for i := 0; i < fieldCount; i++ {
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

		if utils.CompatibleType(valueField, typeOfLazyProc) {
			// LazyProc only supports loading by name
			proc := ld.NewProc(procName)
			// try to load the proc now
			err = proc.Find()
			if err != nil {
				errorOccurred = true
				syntheticErr = multierror.Append(syntheticErr, err)
				continue
			}

			utils.Set(valueField, proc)
		} else if utils.CompatibleType(valueField, typeOfProc) {
			// Windows specific: ordinal
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

			utils.Set(valueField, proc)
		}
	}

	if errorOccurred {
		return syntheticErr
	}
	return nil
}
