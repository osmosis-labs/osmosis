package osmoutils

import "reflect"

func MakeNew[T any]() T {
	var v T
	if typ := reflect.TypeOf(v); typ.Kind() == reflect.Ptr {
		elem := typ.Elem()
		//nolint:forcetypeassert
		return reflect.New(elem).Interface().(T) // must use reflect
	} else {
		return *new(T) // v is not ptr, alloc with new
	}
}
