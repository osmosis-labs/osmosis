package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyBlocksPerEpoch = []byte("BlocksPerEpoch")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(blocksPerEpoch int64) Params {
	return Params{
		BlocksPerEpoch: blocksPerEpoch,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		BlocksPerEpoch: 10,
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateBlocksPerEpoch(p.BlocksPerEpoch); err != nil {
		return err
	}

	return nil

}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyBlocksPerEpoch, &p.BlocksPerEpoch, validateBlocksPerEpoch),
	}
}

func validateBlocksPerEpoch(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 1 {
		return fmt.Errorf("invalid blocks per epoch value: %+v", i)
	}

	return nil
}
