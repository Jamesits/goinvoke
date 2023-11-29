package utils

import "reflect"

// GetStructTag returns the value of a named tag of a struct member
func GetStructTag(f reflect.StructField, tagName string) string {
	return f.Tag.Get(tagName)
}

func CompatibleType(v reflect.Value, t reflect.Type) bool {
	return t.AssignableTo(v.Type()) || v.CanConvert(t)
}

func Set(v reflect.Value, obj any) {
	// https://stackoverflow.com/a/53110731
	v.Set(reflect.ValueOf(obj).Convert(v.Type()))
}
