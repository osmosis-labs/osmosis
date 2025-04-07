package types

import (
	fmt "fmt"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
)

// Incentives parameters key store.
var (
	KeyDistrEpochIdentifier = []byte("DistrEpochIdentifier")
	KeyGroupCreationFee     = []byte("GroupCreationFee")
	KeyCreatorWhitelist     = []byte("CreatorWhitelist")
	KeyInternalUptime       = []byte("InternalUptime")
	KeyMinValueForDistr     = []byte("MinValueForDistr")

	// 100 OSMO
	DefaultGroupCreationFee = sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100_000_000)))
)

// ParamKeyTable returns the key table for the incentive module's parameters.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams takes an epoch distribution identifier and group creation fee, then returns an incentives Params struct.
func NewParams(distrEpochIdentifier string, groupCreationFee sdk.Coins, internalUptime time.Duration, minValueForDistr sdk.Coin) Params {
	return Params{
		DistrEpochIdentifier:         distrEpochIdentifier,
		GroupCreationFee:             groupCreationFee,
		UnrestrictedCreatorWhitelist: []string{},
		InternalUptime:               internalUptime,
		MinValueForDistribution:      minValueForDistr,
	}
}

// DefaultParams returns the default incentives module parameters.
func DefaultParams() Params {
	return Params{
		DistrEpochIdentifier:         "week",
		GroupCreationFee:             DefaultGroupCreationFee,
		UnrestrictedCreatorWhitelist: []string{},
		InternalUptime:               DefaultConcentratedUptime,
		MinValueForDistribution:      DefaultMinValueForDistr,
	}
}

// Validate checks that the incentives module parameters are valid.
func (p Params) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierInterface(p.DistrEpochIdentifier); err != nil {
		return err
	}

	if err := ValidateGroupCreaionFee(p.GroupCreationFee); err != nil {
		return err
	}

	if err := osmoutils.ValidateAddressList(p.UnrestrictedCreatorWhitelist); err != nil {
		return err
	}

	if err := ValidateInternalUptime(p.InternalUptime); err != nil {
		return err
	}

	if err := ValidateMinValueForDistr(p.MinValueForDistribution); err != nil {
		return err
	}

	return nil
}

func ValidateGroupCreaionFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return v.Validate()
}

func ValidateGroupCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return v.Validate()
}

func ValidateInternalUptime(i interface{}) error {
	internalUptime, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	supported := false
	for _, supportedUptime := range cltypes.SupportedUptimes {
		if internalUptime == supportedUptime {
			supported = true

			// We break here to save on iterations
			break
		}
	}

	if !supported {
		return cltypes.UptimeNotSupportedError{Uptime: internalUptime}
	}

	return nil
}

func ValidateMinValueForDistr(i interface{}) error {
	_, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

// ParamSetPairs takes the parameter struct and associates the paramsubspace key and field of the parameters as a KVStore.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistrEpochIdentifier, &p.DistrEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramtypes.NewParamSetPair(KeyGroupCreationFee, &p.GroupCreationFee, ValidateGroupCreaionFee),
		paramtypes.NewParamSetPair(KeyCreatorWhitelist, &p.UnrestrictedCreatorWhitelist, osmoutils.ValidateAddressList),
		paramtypes.NewParamSetPair(KeyInternalUptime, &p.InternalUptime, ValidateInternalUptime),
		paramtypes.NewParamSetPair(KeyMinValueForDistr, &p.MinValueForDistribution, ValidateMinValueForDistr),
	}
}
