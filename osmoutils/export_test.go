package osmoutils

import db "github.com/cometbft/cometbft-db"

func GatherValuesFromIterator[T any](iterator db.Iterator, parseValue func([]byte) (T, error), stopFn func([]byte) bool) ([]T, error) {
	return gatherValuesFromIterator(iterator, parseValue, stopFn)
}

func NoStopFn(key []byte) bool {
	return noStopFn(key)
}
