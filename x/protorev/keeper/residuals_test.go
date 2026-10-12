package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	"github.com/osmosis-labs/osmosis/v31/x/protorev/types"
)

const residualPoolLiquidity = 1_000_000_000_000

// setupResidualPool creates a balancer pool between denomA and denomB and registers it in the
// highest liquidity pool store under (baseDenom, matchDenom).
func (s *KeeperTestSuite) setupResidualPool(baseDenom, matchDenom string) uint64 {
	poolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewCoin(baseDenom, osmomath.NewInt(residualPoolLiquidity)),
		sdk.NewCoin(matchDenom, osmomath.NewInt(residualPoolLiquidity)),
	)
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, baseDenom, matchDenom, poolId)
	return poolId
}

func (s *KeeperTestSuite) mintToProtorev(coins sdk.Coins) {
	s.Require().NoError(s.App.BankKeeper.MintCoins(s.Ctx, types.ModuleName, coins))
}

func (s *KeeperTestSuite) protorevBalance() sdk.Coins {
	return s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
}

func (s *KeeperTestSuite) burnAddressOsmo() osmomath.Int {
	return s.App.BankKeeper.GetBalance(s.Ctx, types.DefaultNullAddress, types.OsmosisDenomination).Amount
}

func (s *KeeperTestSuite) TestBurnResidualBalances() {
	// Base denoms from setupParams: OSMO, Atom, ibc/0CD3...
	tests := []struct {
		name string
		// setup registers pools and mints balances; returns the coins expected to remain in the module.
		setup         func() sdk.Coins
		expectBurnPos bool
	}{
		{
			name: "direct OSMO route",
			setup: func() sdk.Coins {
				s.setupResidualPool(types.OsmosisDenomination, "foo")
				s.mintToProtorev(sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(1_000_000))))
				return sdk.NewCoins()
			},
			expectBurnPos: true,
		},
		{
			name: "two hop route through another base denom",
			setup: func() sdk.Coins {
				s.setupResidualPool("Atom", "bar")
				s.setupResidualPool(types.OsmosisDenomination, "Atom")
				s.mintToProtorev(sdk.NewCoins(sdk.NewCoin("bar", osmomath.NewInt(1_000_000))))
				return sdk.NewCoins()
			},
			expectBurnPos: true,
		},
		{
			name: "denom without a stored pool is left in place",
			setup: func() sdk.Coins {
				spam := sdk.NewCoins(sdk.NewCoin("spam", osmomath.NewInt(1_000_000)))
				s.mintToProtorev(spam)
				return spam
			},
		},
		{
			name: "pair stored under a removed base denom is ignored",
			setup: func() sdk.Coins {
				s.setupResidualPool("removed", "baz")
				coins := sdk.NewCoins(sdk.NewCoin("baz", osmomath.NewInt(1_000_000)))
				s.mintToProtorev(coins)
				return coins
			},
		},
		{
			name: "base denom balances are not swapped",
			setup: func() sdk.Coins {
				s.setupResidualPool(types.OsmosisDenomination, "Atom")
				coins := sdk.NewCoins(
					sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1_000_000)),
					sdk.NewCoin("Atom", osmomath.NewInt(1_000_000)),
				)
				s.mintToProtorev(coins)
				return coins
			},
		},
		{
			name: "failed swap is skipped and other denoms still burn",
			setup: func() sdk.Coins {
				fooPool := s.setupResidualPool(types.OsmosisDenomination, "foo")
				// Point "broken" at a pool that does not contain it so the swap errors.
				s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, types.OsmosisDenomination, "broken", fooPool)
				broken := sdk.NewCoins(sdk.NewCoin("broken", osmomath.NewInt(1_000_000)))
				s.mintToProtorev(broken.Add(sdk.NewCoin("foo", osmomath.NewInt(1_000_000))))
				return broken
			},
			expectBurnPos: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			expectedRemaining := tc.setup()
			burnBefore := s.burnAddressOsmo()

			baseDenoms, err := s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
			s.Require().NoError(err)
			burned := s.App.ProtoRevKeeper.BurnResidualBalances(s.Ctx, baseDenoms)

			s.Require().Equal(types.OsmosisDenomination, burned.Denom)
			s.Require().Equal(tc.expectBurnPos, burned.IsPositive())
			s.Require().Equal(burned.Amount, s.burnAddressOsmo().Sub(burnBefore))
			s.Require().Equal(expectedRemaining, s.protorevBalance())
		})
	}
}

// TestBurnResidualBalancesCap checks that at most MaxResidualSwapsPerEpoch swaps run per call and
// that later calls drain the remainder.
func (s *KeeperTestSuite) TestBurnResidualBalancesCap() {
	s.SetupTest()

	numDenoms := types.MaxResidualSwapsPerEpoch + 5
	for i := 0; i < numDenoms; i++ {
		denom := fmt.Sprintf("residual%d", i)
		s.setupResidualPool(types.OsmosisDenomination, denom)
		s.mintToProtorev(sdk.NewCoins(sdk.NewCoin(denom, osmomath.NewInt(1_000_000))))
	}

	baseDenoms, err := s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)

	s.App.ProtoRevKeeper.BurnResidualBalances(s.Ctx, baseDenoms)
	s.Require().Len(s.protorevBalance(), numDenoms-types.MaxResidualSwapsPerEpoch)

	s.App.ProtoRevKeeper.BurnResidualBalances(s.Ctx, baseDenoms)
	s.Require().Empty(s.protorevBalance())
}

// TestAfterEpochEndBurnsResidualsWithoutDeveloperFee checks that residual balances are burned in full
// by the epoch hook and never pay the developer fee.
func (s *KeeperTestSuite) TestAfterEpochEndBurnsResidualsWithoutDeveloperFee() {
	s.SetupTest()

	devAccount := apptesting.CreateRandomAccounts(1)[0]
	s.App.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, devAccount)

	s.setupResidualPool(types.OsmosisDenomination, "foo")
	s.mintToProtorev(sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(1_000_000))))
	burnBefore := s.burnAddressOsmo()

	err := s.App.ProtoRevKeeper.EpochHooks().AfterEpochEnd(s.Ctx, "day", 1)
	s.Require().NoError(err)

	s.Require().True(s.burnAddressOsmo().GT(burnBefore))
	s.Require().Empty(s.App.BankKeeper.GetAllBalances(s.Ctx, devAccount))
	s.Require().Empty(s.protorevBalance())
}
