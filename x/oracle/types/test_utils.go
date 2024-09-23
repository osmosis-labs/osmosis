package types

import (
	"context"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"
	"math"
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cometbft/cometbft/crypto/secp256k1"
)

const OracleDecPrecision = 8

func GenerateRandomTestCase() (rates []float64, valValAddrs []sdk.ValAddress, stakingKeeper DummyStakingKeeper) {
	valValAddrs = []sdk.ValAddress{}
	mockValidators := []stakingtypes.Validator{}

	base := math.Pow10(OracleDecPrecision)

	r := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	numInputs := 10 + (r.Int() % 100)
	for i := 0; i < numInputs; i++ {
		rate := float64(int64(r.Float64()*base)) / base
		rates = append(rates, rate)

		pubKey := secp256k1.GenPrivKey().PubKey()
		valValAddr := sdk.ValAddress(pubKey.Address())
		valValAddrs = append(valValAddrs, valValAddr)

		power := r.Int63()%1000 + 1
		mockValidator := NewMockValidator(valValAddr, power)
		mockValidators = append(mockValidators, mockValidator)
	}

	stakingKeeper = NewDummyStakingKeeper(mockValidators)

	return
}

var _ StakingKeeper = DummyStakingKeeper{}

// DummyStakingKeeper dummy staking keeper to test ballot
type DummyStakingKeeper struct {
	validators []stakingtypes.Validator
}

// NewDummyStakingKeeper returns new DummyStakingKeeper instance
func NewDummyStakingKeeper(validators []stakingtypes.Validator) DummyStakingKeeper {
	return DummyStakingKeeper{
		validators: validators,
	}
}

func (sk DummyStakingKeeper) GetValidator(ctx context.Context, address sdk.ValAddress) (stakingtypes.Validator, error) {
	for _, validator := range sk.validators {
		if validator.GetOperator() == address.String() {
			return validator, nil
		}
	}
	return stakingtypes.Validator{}, fmt.Errorf("validator not found")
}

func (sk DummyStakingKeeper) TotalBondedTokens(ctx context.Context) (sdkmath.Int, error) {
	return osmomath.ZeroInt(), nil
}

func (sk DummyStakingKeeper) Slash(ctx context.Context, address sdk.ConsAddress, i int64, i2 int64, dec osmomath.Dec) (sdkmath.Int, error) {
	return osmomath.ZeroInt(), nil
}

func (sk DummyStakingKeeper) Jail(ctx context.Context, address sdk.ConsAddress) error {
	return nil
}

func (sk DummyStakingKeeper) ValidatorsPowerStoreIterator(ctx context.Context) (storetypes.Iterator, error) {
	return storetypes.KVStoreReversePrefixIterator(nil, nil), nil
}

func (sk DummyStakingKeeper) MaxValidators(ctx context.Context) (uint32, error) {
	return 100, nil
}

func (sk DummyStakingKeeper) PowerReduction(ctx context.Context) sdkmath.Int {
	return sdk.DefaultPowerReduction
}

func (sk DummyStakingKeeper) Validators() []stakingtypes.Validator {
	return sk.validators
}

func (sk DummyStakingKeeper) GetLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress) int64 {
	val, err := sk.GetValidator(ctx, operator)
	if err != nil {
		return 0
	}
	return val.GetConsensusPower(sdk.DefaultPowerReduction)
}

func NewMockValidator(valAddr sdk.ValAddress, power int64) stakingtypes.Validator {
	return stakingtypes.Validator{
		Status:          stakingtypes.Bonded,
		OperatorAddress: valAddr.String(),
		Tokens:          sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction),
	}
}
