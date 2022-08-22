package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	errNilEpochProvisions      = errors.New("epoch provisions was nil in genesis")
	errNegativeEpochProvisions = errors.New("epoch provisions should be non-negative")
)

// NewMinter returns a new Minter object with the given epoch
// provisions values.
func NewMinter(epochProvisions sdk.Dec) Minter {
	return Minter{
		EpochProvisions: epochProvisions,
	}
}

// InitialMinter returns an initial Minter object.
func InitialMinter() Minter {
	return NewMinter(sdk.NewDec(0))
}

// DefaultInitialMinter returns a default initial Minter object for a new chain.
func DefaultInitialMinter() Minter {
	return InitialMinter()
}

// Validate validates minter. Returns nil on success, error otherewise.
func (m Minter) Validate() error {
	if m.EpochProvisions.IsNil() {
		return errNilEpochProvisions
	}

	if m.EpochProvisions.IsNegative() {
		return errNegativeEpochProvisions
	}
	return nil
}

// NextEpochProvisions returns the epoch provisions.
func (m Minter) NextEpochProvisions(params Params) sdk.Dec {
	return m.EpochProvisions.Mul(params.ReductionFactor)
}

// InflationProvision returns the provisions for a block based on the epoch
// provisions rate. These provisions are first minted from
// the mint module account and then disttributed to all proportions,
// excludeing developer rewards as they are handled by the developer
// vesting module account.
// TODO: test
func (m Minter) InflationProvisions(params Params) sdk.DecCoin {
	provisionAmt := m.EpochProvisions.Mul(sdk.OneDec().Sub(params.DistributionProportions.DeveloperRewards))
	return sdk.NewDecCoinFromDec(params.MintDenom, provisionAmt)
}

// DeveloperVestingEpochProvisions returns the provisions for a block based on the epoch
// provisions rate. These are not minted and distributed from the developer vesting module
// account. These only include developer rewards as all other inflation proportions
// are handled by the mint module account.
// TODO: test
func (m Minter) DeveloperVestingEpochProvisions(params Params) sdk.DecCoin {
	provisionAmt := m.EpochProvisions.Mul(params.DistributionProportions.DeveloperRewards)
	return sdk.NewDecCoinFromDec(params.MintDenom, provisionAmt)
}
