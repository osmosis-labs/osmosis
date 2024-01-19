package wasmbinding

import "github.com/cosmos/cosmos-sdk/codec"

func SetWhitelistedQuery[T any, PT protoTypeG[T]](queryPath string, protoType PT) {
	setWhitelistedQuery(queryPath, protoType)
}

func GetWhitelistedQuery(queryPath string) (codec.ProtoMarshaler, error) {
	return getWhitelistedQuery(queryPath)
}
