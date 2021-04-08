package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMinter returns a new Minter object with the given annual
// provisions values.
func NewMinter(annualProvisions sdk.Dec) Minter {
	return Minter{
		AnnualProvisions: annualProvisions,
	}
}

// InitialMinter returns an initial Minter object
func InitialMinter() Minter {
	return NewMinter(sdk.NewDec(0))
}

// DefaultInitialMinter returns a default initial Minter object for a new chain
func DefaultInitialMinter() Minter {
	return InitialMinter()
}

// validate minter
func ValidateMinter(minter Minter) error {
	return nil
}

// NextAnnualProvisions returns the annual provisions
func (m Minter) NextAnnualProvisions(params Params) sdk.Dec {
	return m.AnnualProvisions.Mul(params.ReductionFactorForEvent)
}

// EpochProvision returns the provisions for a block based on the annual
// provisions rate.
func (m Minter) EpochProvision(params Params) sdk.Coin {
	provisionAmt := m.AnnualProvisions.QuoInt(sdk.NewInt(int64(params.EpochsPerYear)))
	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}
