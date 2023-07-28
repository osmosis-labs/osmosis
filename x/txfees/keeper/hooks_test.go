package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v17/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v17/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v17/x/txfees/types"
)

var defaultPooledAssetAmount = int64(500)

func (s *KeeperTestSuite) preparePool(denom string) (poolID uint64, pool poolmanagertypes.PoolI) {
	baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	poolID = s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(denom, defaultPooledAssetAmount),
	)
	pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolID)
	s.Require().NoError(err)
	err = s.ExecuteUpgradeFeeTokenProposal(denom, poolID)
	s.Require().NoError(err)
	return poolID, pool
}

func (s *KeeperTestSuite) TestTxFeesAfterEpochEnd() {
	s.SetupTest(false)
	baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)

	// create pools for three separate fee tokens
	uion := "uion"
	_, uionPool := s.preparePool(uion)
	atom := "atom"
	_, atomPool := s.preparePool(atom)
	ust := "ust"
	_, ustPool := s.preparePool(ust)

	tests := []struct {
		name         string
		coins        sdk.Coins
		baseDenom    string
		denoms       []string
		poolTypes    []poolmanagertypes.PoolI
		spreadFactor sdk.Dec
		expectPass   bool
	}{
		{
			name:         "One non-osmo fee token (uion): TxFees AfterEpochEnd",
			coins:        sdk.Coins{sdk.NewInt64Coin(uion, 10)},
			baseDenom:    baseDenom,
			denoms:       []string{uion},
			poolTypes:    []poolmanagertypes.PoolI{uionPool},
			spreadFactor: sdk.MustNewDecFromStr("0"),
		},
		{
			name:         "Multiple non-osmo fee token: TxFees AfterEpochEnd",
			coins:        sdk.Coins{sdk.NewInt64Coin(atom, 20), sdk.NewInt64Coin(ust, 30)},
			baseDenom:    baseDenom,
			denoms:       []string{atom, ust},
			poolTypes:    []poolmanagertypes.PoolI{atomPool, ustPool},
			spreadFactor: sdk.MustNewDecFromStr("0"),
		},
	}

	finalOutputAmount := sdk.NewInt(0)

	for _, tc := range tests {
		tc := tc

		s.Run(tc.name, func() {
			for i, coin := range tc.coins {
				// Get the output amount in osmo denom
				pool, ok := tc.poolTypes[i].(gammtypes.CFMMPoolI)
				s.Require().True(ok)

				expectedOutput, err := pool.CalcOutAmtGivenIn(s.Ctx,
					sdk.Coins{sdk.Coin{Denom: tc.denoms[i], Amount: coin.Amount}},
					tc.baseDenom,
					tc.spreadFactor)
				s.NoError(err)
				// sanity check for the expectedAmount
				s.True(coin.Amount.GTE(expectedOutput.Amount))

				finalOutputAmount = finalOutputAmount.Add(expectedOutput.Amount)

				// Deposit some fee amount (non-native-denom) to the fee module account
				_, _, addr0 := testdata.KeyTestPubAddr()
				err = simapp.FundAccount(s.App.BankKeeper, s.Ctx, addr0, sdk.Coins{coin})
				s.NoError(err)
				err = s.App.BankKeeper.SendCoinsFromAccountToModule(s.Ctx, addr0, types.NonNativeFeeCollectorName, sdk.Coins{coin})
				s.NoError(err)
			}

			// checks the balance of the non-native denom in module account
			moduleAddrNonNativeFee := s.App.AccountKeeper.GetModuleAddress(types.NonNativeFeeCollectorName)
			s.Equal(s.App.BankKeeper.GetAllBalances(s.Ctx, moduleAddrNonNativeFee), tc.coins)

			// End of epoch, so all the non-osmo fee amount should be swapped to osmo and transfer to fee module account
			params := s.App.IncentivesKeeper.GetParams(s.Ctx)
			futureCtx := s.Ctx.WithBlockTime(time.Now().Add(time.Minute))
			err := s.App.TxFeesKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, int64(1))
			s.NoError(err)

			// check the balance of the native-basedenom in module
			moduleAddrFee := s.App.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
			moduleBaseDenomBalance := s.App.BankKeeper.GetBalance(s.Ctx, moduleAddrFee, tc.baseDenom)

			// non-osmos module account should be empty as all the funds should be transferred to osmo module
			s.Empty(s.App.BankKeeper.GetAllBalances(s.Ctx, moduleAddrNonNativeFee))
			// check that the total osmo amount has been transferred to module account
			s.Equal(moduleBaseDenomBalance.Amount, finalOutputAmount)
		})
	}
}
