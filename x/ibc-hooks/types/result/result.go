package result

import (
	"C"
	"encoding/base64"
)

var (
	Ok           byte = 0
	QueryError   byte = 1
	ExecuteError byte = 2
)

func markError(code byte, data []byte) []byte {
	return append([]byte{code}, data...)
}

func markOk(data []byte) []byte {
	return append([]byte{Ok}, data...)
}

func EncodeResultFromError(code byte, err error) string {
	marked := markError(code, []byte(err.Error()))
	return base64.StdEncoding.EncodeToString(marked)
}

func EncodeResultFromOk(data []byte) string {
	marked := markOk(data)
	return base64.StdEncoding.EncodeToString(marked)
}
