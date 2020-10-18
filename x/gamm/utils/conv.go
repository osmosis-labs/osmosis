package utils

import "encoding/binary"

func Uint64ToBytes(i uint64) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, i)
	return key
}
