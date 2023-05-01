package osmoutils

import (
	db "github.com/tendermint/tm-db"
	"golang.org/x/exp/constraints"
)

func GatherValuesFromIterator[T any](iterator db.Iterator, parseValue func([]byte) (T, error), stopFn func([]byte) bool) ([]T, error) {
	return gatherValuesFromIterator(iterator, parseValue, stopFn)
}

func NoStopFn(key []byte) bool {
	return noStopFn(key)
}

func IThSmallest[T constraints.Ordered](s []T, i int, less LessFunc[T]) T {
	return iThSmallest(s, i, less)
}
