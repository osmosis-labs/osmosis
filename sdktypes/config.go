package sdktypes

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultSDKConfig() *sdk.Config {
	sdkConfig := sdk.NewConfig()
	return sdkConfig
}
