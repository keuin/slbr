package common

type Opt[T any] struct {
	thing T
	err   error
}

func Optional[T any](thing T, err error) Opt[T] {
	return Opt[T]{
		thing: thing,
		err:   err,
	}
}

func (o Opt[T]) OrElse(thing T) T {
	if o.err != nil {
		return thing
	}
	return o.thing
}
