package goinvoke

import "errors"

var (
	ErrorNotFound        = errors.New("not found")
	ErrorUnmarshalFailed = errors.New("unmarshal failed")
)
