package transmuter_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	incentivetypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
)

const (
	denomA = apptesting.DefaultTransmuterDenomA
	denomB = apptesting.DefaultTransmuterDenomB
)

// Suite for the transmuter contract.
type TransmuterSuite struct {
	apptesting.KeeperTestHelper
}

var (
	defaultPoolId       = uint64(1)
	defaultAmount       = osmomath.NewInt(100)
	initalDefaultSupply = sdk.NewCoins(sdk.NewCoin(denomA, defaultAmount), sdk.NewCoin(denomB, defaultAmount))
	uosmo               = appparams.BaseCoinUnit

	defaultDenoms = []string{denomA, denomB}
)

func TestTransmuterSuite(t *testing.T) {
	suite.Run(t, new(TransmuterSuite))
}

// This test functionally tests that the transmuter contract works as expected.
// It validates:
// - LP and tokenfactory share creation
// - Ability to lock, create gauges for and incentivize such shares
func (s *TransmuterSuite) TestFunctionalTransmuter() {
	s.Setup()

	const (
		exppectedDenomPrefix = tokenfactorytypes.ModuleDenomPrefix + "/"
		expectedDenomSuffix  = "/transmuter/poolshare"
	)

	// Set base denom
	s.App.IncentivesKeeper.SetParam(s.Ctx, incentivetypes.KeyMinValueForDistr, sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10000)))

	// Create Transmuter pool
	transmuter := s.PrepareCosmWasmPool()

	contractAddress := transmuter.GetContractAddress()
	expectedShareDenom := exppectedDenomPrefix + contractAddress + expectedDenomSuffix

	// Validate that tokenfactory denom is created  in the desired format.
	denomIteraror := s.App.TokenFactoryKeeper.GetAllDenomsIterator(s.Ctx)
	defer denomIteraror.Close()
	s.Require().True(denomIteraror.Valid())
	s.Require().Equal(expectedShareDenom, string(denomIteraror.Value()))

	// Fund account
	s.FundAcc(s.TestAccs[0], initalDefaultSupply)

	// Join pool
	s.JoinTransmuterPool(s.TestAccs[0], defaultPoolId, initalDefaultSupply)

	// Check that the number of shares equals to the sum of the initial token amounts.
	shareCoin := s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[0], expectedShareDenom)
	s.Require().Equal(defaultAmount.MulRaw(2), shareCoin.Amount)

	// Attempt to incentivize the tokenfactory shares

	// Lock shares
	shareCoins := sdk.NewCoins(shareCoin)
	lockDuration := time.Hour
	_, err := s.App.LockupKeeper.CreateLock(s.Ctx, s.TestAccs[0], shareCoins, lockDuration)
	s.Require().NoError(err)

	// Create gauge
	incentive := sdk.NewCoins(sdk.NewCoin(uosmo, osmomath.NewInt(1_000_000)))
	s.FundAcc(s.TestAccs[1], incentive)
	gaugeId, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, true, s.TestAccs[1], incentive, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         expectedShareDenom,
		Duration:      lockDuration,
	}, s.Ctx.BlockTime(), 1, 0)
	s.Require().NoError(err)
	gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
	s.Require().NoError(err)

	// Distribute rewards
	coins, err := s.App.IncentivesKeeper.Distribute(s.Ctx, []incentivetypes.Gauge{*gauge})
	s.Require().NoError(err)

	// Confirm that the rewards are distributed with no errors.
	s.Require().Equal(incentive.String(), coins.String())
}
