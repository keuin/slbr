package common

import (
	"fmt"
)

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
