package osmolbp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v043_temp/address"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "osmolbp"

	// RouterKey is the message route for authz
	RouterKey = ModuleName

	// QuerierRoute is the querier route for authz
	QuerierRoute = ModuleName
)

func NewLbpAddress(lbpId uint64) sdk.AccAddress {
	key := append([]byte("lbp"), sdk.Uint64ToBigEndian(lbpId)...)
	return address.Module(ModuleName, key)
}
