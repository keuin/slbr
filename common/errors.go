package common

import (
	"errors"
	"fmt"
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

/*
Task errors.
*/

type RecoverableTaskError struct {
	err     error
	message string
}

func (e *RecoverableTaskError) Error() string {
	return fmt.Sprintf("%v: %v", e.message, e.err)
}

func (e *RecoverableTaskError) Unwrap() error {
	return e.err
}

func NewRecoverableTaskError(message string, err error) error {
	return &RecoverableTaskError{message: message, err: err}
}

type UnrecoverableTaskError struct {
	err     error
	message string
}

func (e *UnrecoverableTaskError) Error() string {
	return fmt.Sprintf("%v: %v", e.message, e.err)
}

func (e *UnrecoverableTaskError) Unwrap() error {
	return e.err
}

func NewUnrecoverableTaskError(message string, err error) error {
	return &UnrecoverableTaskError{message: message, err: err}
}
