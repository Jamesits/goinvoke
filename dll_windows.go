//go:build windows

package goinvoke

import (
	"github.com/jamesits/goinvoke/utils"
	"golang.org/x/sys/windows"
	"reflect"
)

var typeOfLazyProc = reflect.TypeOf((*windows.LazyProc)(nil))
var typeOfProc = reflect.TypeOf((*windows.Proc)(nil))

// convert a LazyDLL to DLL, assume it has been loaded.
func unLazy(lazyDLL *windows.LazyDLL) *windows.DLL {
	return &windows.DLL{
		Name:   lazyDLL.Name,
		Handle: windows.Handle(lazyDLL.Handle()),
	}
}

func newLazyDLL(path string) *windows.LazyDLL {
	if utils.IsImplicitRelativePath(path) {
		return windows.NewLazySystemDLL(path)
	} else {
		return windows.NewLazyDLL(path)
	}
}
