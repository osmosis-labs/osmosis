package utils

import (
	"encoding/binary"
	"strconv"
)

func Uint64ToBytes(i uint64) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, i)
	return key
}

func Uint64ToString(i uint64) string {
	return strconv.FormatUint(i, 10)
}
