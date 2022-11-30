package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
)

var (
	ErrNoOsmoPool              = sdkerrors.Register(types.ModuleName, 100, "There is no Osmo pool for the given denom")
	ErrNoAtomPool              = sdkerrors.Register(types.ModuleName, 101, "There is no Osmo pool for the given denom")
	ErrNoAtomRoute             = sdkerrors.Register(types.ModuleName, 102, "There is no Atom route for the given denom")
	ErrNoOsmoRoute             = sdkerrors.Register(types.ModuleName, 103, "There is no Osmo route for the given denom")
	ErrNoTokenPairRoutes       = sdkerrors.Register(types.ModuleName, 104, "No hot route exists for the given pool id")
	ErrNoProtoRevTrades        = sdkerrors.Register(types.ModuleName, 105, "There are no proto-rev trades for the module")
	ErrNoProtoRevProfitByDenom = sdkerrors.Register(types.ModuleName, 106, "There is no proto-rev profit for the given denom")
	ErrNoProtoRevTradesByPool  = sdkerrors.Register(types.ModuleName, 107, "There are no proto-rev trades for the given pool id")
	ErrHotRoutesNotConfigured  = sdkerrors.Register(types.ModuleName, 108, "Hot routes were not set in the initialization of the module")
	ErrProtoRevNotConfigured   = sdkerrors.Register(types.ModuleName, 109, "Not configured for the protorev module")
	ErrNoAdminAccount          = sdkerrors.Register(types.ModuleName, 110, "No admin account has been set")
	ErrNoDeveloperAccount      = sdkerrors.Register(types.ModuleName, 111, "No developer account has been set")
	ErrUnauthorized            = sdkerrors.Register(types.ModuleName, 112, "Unauthorized transaction")
)
