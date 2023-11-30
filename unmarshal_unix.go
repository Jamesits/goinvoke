package goinvoke

import (
	"errors"
	"github.com/hashicorp/go-multierror"
	"github.com/jamesits/goinvoke/utils"
	"reflect"
)

// Unmarshal loads the DLL into memory, then fills all struct fields with type *windows.LazyProc with exported functions.
func Unmarshal(path string, v any) error {
	var err error
	var syntheticErr = errors.New("unmarshal failed")
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
