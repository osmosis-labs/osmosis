package osmoutils

import db "github.com/tendermint/tm-db"

func GatherValuesFromIteratorWithStop[T any](iterator db.Iterator, parseValue func([]byte) (T, error), stopFn func([]byte) bool) ([]T, error) {
	return gatherValuesFromIteratorWithStop(iterator, parseValue, stopFn)
}
