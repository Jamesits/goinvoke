package goinvoke

type FunctionPointer interface {
	Addr() uintptr
	Call(...uintptr) (uintptr, uintptr, error)
}
