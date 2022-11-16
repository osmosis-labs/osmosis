package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"
)

var (
	ErrNoOsmoPool           = sdkerrors.Register(types.ModuleName, 100, "There is no Osmo pool for the given denom")
	ErrNoAtomPool           = sdkerrors.Register(types.ModuleName, 101, "There is no Osmo pool for the given denom")
	ErrNoRoute              = sdkerrors.Register(types.ModuleName, 103, "No route exists for the given pair of denoms")
	ErrNoProtoRevStatistics = sdkerrors.Register(types.ModuleName, 104, "There are no proto-rev statistics for the module")
)
