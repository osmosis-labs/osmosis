package types

import (
	"github.com/osmosis-labs/osmosis/osmoutils"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyWhitelistedFeeTokenSetters = []byte("WhitelistedFeeTokenSetters")
)

// ParamTable for txfees module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(whitelistedFeeTokenSetters []string) Params {
	return Params{
		WhitelistedFeeTokenSetters: whitelistedFeeTokenSetters,
	}
}

// DefaultParams are the default txfees module parameters.
func DefaultParams() Params {
	return Params{
		WhitelistedFeeTokenSetters: []string{},
	}
}

// validate params.
func (p Params) Validate() error {
	if err := osmoutils.ValidateAddressList(p.WhitelistedFeeTokenSetters); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyWhitelistedFeeTokenSetters, &p.WhitelistedFeeTokenSetters, osmoutils.ValidateAddressList),
	}
}
