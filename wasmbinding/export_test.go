package wasmbinding

import "github.com/cosmos/cosmos-sdk/codec"

func SetWhitelistedQuery(queryPath string, factoryFunc func() codec.ProtoMarshaler) {
	setWhitelistedQuery(queryPath, factoryFunc)
}
