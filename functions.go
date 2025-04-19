package gocache

import (
	"reflect"
	"strings"
)

func GetTypeName[T any]() string {
	return strings.ToLower(reflect.TypeOf((*T)(nil)).Elem().Name())
}
