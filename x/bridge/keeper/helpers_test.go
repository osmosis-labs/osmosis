// This file contains all helpers for the keeper_test package.

package keeper_test

import (
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

var (
	pk1        = ed25519.GenPrivKey().PubKey()
	addr1Bytes = sdk.AccAddress(pk1.Address())
	addr1      = addr1Bytes.String()

	pk2        = ed25519.GenPrivKey().PubKey()
	addr2Bytes = sdk.AccAddress(pk2.Address())
	addr2      = addr2Bytes.String()

	assetID1 = types.AssetID{
		SourceChain: types.DefaultBitcoinChainName,
		Denom:       "btc1",
	}
	asset1 = types.Asset{
		Id:       assetID1,
		Status:   types.AssetStatus_ASSET_STATUS_BLOCKED_BOTH,
		Exponent: types.DefaultBitcoinExponent,
	}

	assetID2 = types.AssetID{
		SourceChain: types.DefaultBitcoinChainName,
		Denom:       "btc2",
	}
	asset2 = types.Asset{
		Id:       assetID2,
		Status:   types.AssetStatus_ASSET_STATUS_BLOCKED_BOTH,
		Exponent: types.DefaultBitcoinExponent,
	}
)
