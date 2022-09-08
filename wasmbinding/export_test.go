package wasmbinding

import "github.com/cosmos/cosmos-sdk/codec"

func SetWhitelistedQuery(queryPath string, protoType codec.ProtoMarshaler) {
	setWhitelistedQuery(queryPath, protoType)
}
