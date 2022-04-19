package common

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/osmosis-labs/osmosis/v7/app/params"
)

var (
	EncodingConfig params.EncodingConfig
	Cdc            codec.Codec
)

func init() {
	EncodingConfig, Cdc = InitEncodingConfigAndCdc()
}
