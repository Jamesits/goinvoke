//go:build unix

package goinvoke

import (
	"errors"
	"github.com/hashicorp/go-multierror"
	"github.com/jamesits/goinvoke/utils"
	"reflect"
)

var typeOfLazyProc = reflect.TypeOf((*LazyProc)(nil))
var typeOfProc = reflect.TypeOf((*Proc)(nil))

// Unmarshal loads the DLL into memory, then fills all struct fields with type *windows.LazyProc with exported functions.
func Unmarshal(path string, v any) error {
	var err error
	var syntheticErr = errors.New("unmarshal failed")
	var errorOccurred = false

	var ld *LazyDLL
	if utils.IsImplicitRelativePath(path) {
		ld = NewLazySystemDLL(path)
	} else {
		ld = NewLazyDLL(path)
	}

	err = ld.Load()
	if err != nil {
		errorOccurred = true
		syntheticErr = multierror.Append(syntheticErr, err)
		return syntheticErr
	}

	// create a corresponding windows.Dll object for compatibility
	d := &DLL{
		Name:   ld.Name,
		Handle: ld.Handle(),
	}

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
			proc, err := d.FindProc(procName)
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
