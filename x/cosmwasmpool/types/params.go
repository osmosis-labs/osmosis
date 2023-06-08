package types

import (
	"errors"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// DefaultPoolMigrationLimit is the max number of pools that can be migrated at once.
	// Note that 20 was chosen arbitrarily to have a constant bound on the number of pools migrated.
	DefaultPoolMigrationLimit = 20
)

// Parameter store keys.
var (
	KeyCodeIdWhitelist    = []byte("CodeIdWhitelist")
	KeyPoolMigrationLimit = []byte("PoolMigrationLimit")
)

// ParamTable for cosmwasmpool module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams() Params {
	return Params{
		CodeIdWhitelist: []uint64{},
	}
}

// DefaultParams are the default cosmwasmpool module parameters.
func DefaultParams() Params {
	return Params{
		CodeIdWhitelist:    []uint64{},
		PoolMigrationLimit: DefaultPoolMigrationLimit,
	}
}

// Validate validates params.
func (p Params) Validate() error {
	return validateCodeIdWhitelist(p.CodeIdWhitelist)
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyCodeIdWhitelist, &p.CodeIdWhitelist, validateCodeIdWhitelist),
		paramtypes.NewParamSetPair(KeyPoolMigrationLimit, &p.PoolMigrationLimit, validatePoolMigrationLimit),
	}
}

func validateCodeIdWhitelist(value interface{}) error {
	return nil
}

func validatePoolMigrationLimit(value interface{}) error {
	poolMigrationLimit, ok := value.(uint64)
	if !ok {
		return errors.New("invalid type for pool migration limit")
	}

	if poolMigrationLimit == 0 {
		return errors.New("pool migration limit must be greater than 0")
	}

	return nil
}
