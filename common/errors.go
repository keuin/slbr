package common

import (
	"errors"
	"reflect"
)

// IsErrorOfType is a modified version of errors.Is, which loosen the check condition
func IsErrorOfType(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflect.TypeOf(target).Comparable()
	for {
		if isComparable && reflect.TypeOf(target) == reflect.TypeOf(err) {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		// TODO: consider supporting target.Is(err). This would allow
		// user-definable predicates, but also may allow for coping with sloppy
		// APIs, thereby making it easier to get away with them.
		if err = errors.Unwrap(err); err == nil {
			return false
		}
	}
}
