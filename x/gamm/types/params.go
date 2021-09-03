package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
)

// Parameter store keys
var (
	KeyPoolCreationFee             = []byte("PoolCreationFee")
	KeyNumTwapHistoryPerDeletion   = []byte("NumTwapHistoryPerDeletion")
	KeyTwapHistoryDeletionInterval = []byte("TwapHistoryDeletionInterval")
	KeyTwapHistoryKeepDuration     = []byte("TwapHistoryKeepDuration")
)

// ParamTable for gamm module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(poolCreationFee sdk.Coins) Params {
	return Params{
		PoolCreationFee: poolCreationFee,
	}
}

// default gamm module parameters
func DefaultParams() Params {
	return Params{
		PoolCreationFee: sdk.Coins{sdk.NewInt64Coin(appparams.BaseCoinUnit, 1000_000_000)}, // 1000 OSMO
		// unit of pools per deletion iteration
		NumTwapHistoryPerDeletion: 10,
		// unit of blocks
		TwapHistoryDeletionInterval: 5,
		TwapHistoryKeepDuration:     time.Duration(time.Hour * 24),
	}
}

// validate params
func (p Params) Validate() error {
	if err := validatePoolCreationFee(p.PoolCreationFee); err != nil {
		return err
	}

	return nil

}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPoolCreationFee, &p.PoolCreationFee, validatePoolCreationFee),
		paramtypes.NewParamSetPair(KeyNumTwapHistoryPerDeletion, &p.NumTwapHistoryPerDeletion, validateNumTwapHistoryPerDeletion),
		paramtypes.NewParamSetPair(KeyTwapHistoryDeletionInterval, &p.TwapHistoryDeletionInterval, validateTwapHistoryDeletionInterval),
		paramtypes.NewParamSetPair(KeyTwapHistoryKeepDuration, &p.TwapHistoryKeepDuration, validateTwapHistoryKeepDuration),
	}
}

func validatePoolCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid pool creation fee: %+v", i)
	}

	return nil
}

func validateNumTwapHistoryPerDeletion(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateTwapHistoryDeletionInterval(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invlaid parameter type(value of zero): %T", i)
	}

	return nil
}

func validateTwapHistoryKeepDuration(i interface{}) error {
	_, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
