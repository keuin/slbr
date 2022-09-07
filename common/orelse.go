package common

import "reflect"

type Opt[T any] interface {
	OrElse(thing T) T
}

type OptError[T any] struct {
	thing T
	err   error
}

type OptNull[T any] struct {
	ptr *T
}

type OptZero[T any] struct {
	thing T
}

func (o OptNull[T]) OrElse(thing T) T {
	if o.ptr != nil {
		return *o.ptr
	}
	return thing
}

func Errorable[T any](thing T, err error) Opt[T] {
	return OptError[T]{
		thing: thing,
		err:   err,
	}
}

func (o OptError[T]) OrElse(thing T) T {
	if o.err != nil {
		return thing
	}
	return o.thing
}

func Nullable[T any](ptr *T) Opt[T] {
	return OptNull[T]{
		ptr: ptr,
	}
}

func Zeroable[T any](thing T) Opt[T] {
	return OptZero[T]{
		thing: thing,
	}
}

func (o OptZero[T]) OrElse(thing T) T {
	var zero T
	if reflect.DeepEqual(zero, o.thing) {
		return thing
	}
	return o.thing
}
