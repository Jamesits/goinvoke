package utils

import "reflect"

// GetStructTag returns the value of a named tag of a struct member
func GetStructTag(f reflect.StructField, tagName string) string {
	return f.Tag.Get(tagName)
}
