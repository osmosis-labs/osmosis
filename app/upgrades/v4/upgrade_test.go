package v4_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app"
	v4 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v4"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	app       *app.OsmosisApp
	preModule appmodule.HasPreBlocker
	HomeDir   string
}

func (s *UpgradeTestSuite) SetupTest() {
	s.HomeDir = fmt.Sprintf("%d", rand.Int())
	s.app = app.SetupWithCustomHome(false, s.HomeDir)

	s.ctx = s.app.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
	s.preModule = upgrade.NewAppModule(s.app.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))
}

func (s *UpgradeTestSuite) TearDownTest() {
	os.RemoveAll(s.HomeDir)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const dummyUpgradeHeight = 5

func (s *UpgradeTestSuite) TestUpgradePayments() {
	testCases := []struct {
		msg         string
		pre_update  func()
		update      func()
		post_update func()
		expPass     bool
	}{
		{
			"Test community pool payouts for Prop 12",
			func() {
				// mint coins to distribution module / feepool.communitypool

				bal := int64(1000000000000)
				coin := sdk.NewInt64Coin(appparams.BaseCoinUnit, bal)
				coins := sdk.NewCoins(coin)
				err := s.app.BankKeeper.MintCoins(s.ctx, "mint", coins)
				s.Require().NoError(err)
				err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, "mint", "distribution", coins)
				s.Require().NoError(err)
				feePool, err := s.app.DistrKeeper.FeePool.Get(s.ctx)
				if err != nil {
					panic(err)
				}
				feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinFromCoin(coin))
				err = s.app.DistrKeeper.FeePool.Set(s.ctx, feePool)
				if err != nil {
					panic(err)
				}
			},
			func() {
				// run upgrade
				s.ctx = s.ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v4", Height: dummyUpgradeHeight}
				err := s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, plan)
				s.Require().NoError(err)
				_, err = s.app.UpgradeKeeper.GetUpgradePlan(s.ctx)
				s.Require().NoError(err)

				s.ctx = s.ctx.WithHeaderInfo(header.Info{Height: dummyUpgradeHeight, Time: s.ctx.BlockTime().Add(time.Second)}).WithBlockHeight(dummyUpgradeHeight)
				s.Require().NotPanics(func() {
					_, err := s.preModule.PreBlock(s.ctx)
					s.Require().NoError(err)
				})
			},
			func() {
				total := int64(0)

				// check that each account got the payment expected
				payments := v4.GetProp12Payments()
				for _, payment := range payments {
					addr, err := sdk.AccAddressFromBech32(payment[0])
					s.Require().NoError(err)
					amount, err := strconv.ParseInt(strings.TrimSpace(payment[1]), 10, 64)
					s.Require().NoError(err)
					coin := sdk.NewInt64Coin(appparams.BaseCoinUnit, amount)

					accBal := s.app.BankKeeper.GetBalance(s.ctx, addr, appparams.BaseCoinUnit)
					s.Require().Equal(coin, accBal)

					total += amount
				}

				// check that the total paid out was as expected
				s.Require().Equal(total, int64(367926557424))

				expectedBal := 1000000000000 - total

				// check that distribution module account balance has been reduced correctly
				distAddr := s.app.AccountKeeper.GetModuleAddress("distribution")
				distBal := s.app.BankKeeper.GetBalance(s.ctx, distAddr, appparams.BaseCoinUnit)
				s.Require().Equal(distBal, sdk.NewInt64Coin(appparams.BaseCoinUnit, expectedBal))

				// check that feepool.communitypool has been reduced correctly
				feePool, err := s.app.DistrKeeper.FeePool.Get(s.ctx)
				if err != nil {
					panic(err)
				}
				s.Require().Equal(feePool.GetCommunityPool(), sdk.NewDecCoins(sdk.NewInt64DecCoin(appparams.BaseCoinUnit, expectedBal)))

				// Check that gamm Minimum Fee has been set correctly

				// Kept as comments for recordkeeping. Since SetParams is now private, the changes being tested for can no longer be made:
				//  	gammParams := s.app.GAMMKeeper.GetParams(suite.ctx)
				//  	expectedCreationFee := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.OneInt()))
				//  	s.Require().Equal(gammParams.PoolCreationFee, expectedCreationFee)
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.pre_update()
			tc.update()
			tc.post_update()
		})
	}
}
