package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OracleKeeper for feeder validation
type OracleKeeper interface {
	ValidateFeeder(ctx sdk.Context, feederAddr sdk.AccAddress, validatorAddr sdk.ValAddress) error
}
