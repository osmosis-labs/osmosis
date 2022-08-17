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
		EpochProvisions:          epochProvisions,
		LastTotalInflationAmount: sdk.ZeroDec(),
		LastTotalVestedAmount:    sdk.ZeroDec(),
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

// EpochProvision returns the provisions for a block based on the epoch
// provisions rate. It excludes developer rewards as they are
// handled by the developer vesting module account.
func (m Minter) EpochProvision(params Params) sdk.Coin {
	provisionAmt := m.EpochProvisions.Mul(sdk.OneDec().Sub(params.DistributionProportions.DeveloperRewards)).TruncateInt()
	return sdk.NewCoin(params.MintDenom, provisionAmt)
}

// EpochProvision returns the provisions for a block based on the epoch
// provisions rate. It excludes developer rewards as they are
// handled by the developer vesting module account.
func (m Minter) DeveloperVestingEpochProvision(params Params) sdk.Coin {
	provisionAmt := m.EpochProvisions.Mul(params.DistributionProportions.DeveloperRewards).TruncateInt()
	return sdk.NewCoin(params.MintDenom, provisionAmt)
}
