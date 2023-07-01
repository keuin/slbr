package dmmsg

import (
	"fmt"
	"reflect"
)

func castValue[T any](obj interface{}) (thing T, err error) {
	casted, ok := (obj).(T)
	if !ok {
		err = fmt.Errorf("%s: required value is not of type \"%v\": %v",
			InvalidDanmakuJson, reflect.TypeOf(thing).String(), obj)
		return
	}
	thing = casted
	return
}
