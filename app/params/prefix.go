package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// configurations
var (
	AccountAddressPrefix = "osm1"
	AccountPubKeyPrefix  = "osm1pub"

	ValidatorAddressPrefix = "osm1valoper"
	ValidatorPubKeyPrefix  = "osm1valoperpub"

	ConsNodeAddressPrefix = "osm1valcons"
	ConsNodePubKeyPrefix  = "osm1valconspub"
)

// SetBech32Prefixes set bech32 prefixes
func SetBech32Prefixes() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountPubKeyPrefix)
	config.SetBech32PrefixForValidator(ValidatorAddressPrefix, ValidatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(ConsNodeAddressPrefix, ConsNodePubKeyPrefix)
}
