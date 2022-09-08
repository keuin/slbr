package common

/*
Golang is a piece of shit. Its creators are paranoids.
*/

import (
	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

func Min[T Number](t1 T, t2 T) T {
	if t1 < t2 {
		return t1
	}
	return t2
}

func Max[T Number](t1 T, t2 T) T {
	if t1 > t2 {
		return t1
	}
	return t2
}
