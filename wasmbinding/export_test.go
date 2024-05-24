package wasmbinding

import (
	"github.com/cosmos/gogoproto/proto"
)

func SetWhitelistedQuery[T any, PT protoTypeG[T]](queryPath string, protoType PT) {
	setWhitelistedQuery(queryPath, protoType)
}

func GetWhitelistedQuery(queryPath string) (proto.Message, error) {
	return getWhitelistedQuery(queryPath)
}
