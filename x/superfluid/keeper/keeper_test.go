package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/tendermint/tendermint/crypto/ed25519"

	lockupkeeper "github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"

	epochtypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v7/x/mint/types"
)

type KeeperTestSuite struct {
	suite.Suite
	helper apptesting.KeeperTestHelper

	ctx         sdk.Context
	queryClient types.QueryClient
	app         *app.OsmosisApp
}

func (suite *KeeperTestSuite) GetSuite() *suite.Suite {
	return &suite.Suite
}
func (suite *KeeperTestSuite) GetCtx() sdk.Context {
	return suite.ctx
}
func (suite *KeeperTestSuite) GetApp() *app.OsmosisApp {
	return suite.app
}
func (suite *KeeperTestSuite) SetCtx(ctx sdk.Context) {
	suite.ctx = ctx
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)

	suite.helper = apptesting.KeeperTestHelper{
		Suite: suite.Suite,
		App:   suite.app,
		Ctx:   suite.ctx,
	}
	startTime := time.Unix(1645580000, 0)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: startTime.UTC()})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.SuperfluidKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
	suite.SetupDefaultPool()

	unbondingDuration := suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime

	suite.app.IncentivesKeeper.SetLockableDurations(suite.ctx, []time.Duration{
		time.Hour * 24 * 14,
		time.Hour,
		time.Hour * 3,
		time.Hour * 7,
		unbondingDuration,
	})

	// TODO: Revisit if this is needed, it was added due to another bug in testing that is now fixed.
	epochIdentifier := suite.app.SuperfluidKeeper.GetEpochIdentifier(suite.ctx)
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochtypes.EpochInfo{
		Identifier:              epochIdentifier,
		StartTime:               startTime,
		Duration:                time.Hour,
		CurrentEpochStartTime:   startTime,
		CurrentEpochStartHeight: 1,
		CurrentEpoch:            1,
		EpochCountingStarted:    true,
	})

	mintParams := suite.app.MintKeeper.GetParams(suite.ctx)
	mintParams.EpochIdentifier = epochIdentifier
	mintParams.DistributionProportions = minttypes.DistributionProportions{
		Staking:          sdk.OneDec(),
		PoolIncentives:   sdk.ZeroDec(),
		DeveloperRewards: sdk.ZeroDec(),
		CommunityPool:    sdk.ZeroDec(),
	}
	suite.app.MintKeeper.SetParams(suite.ctx, mintParams)
	suite.app.MintKeeper.SetMinter(suite.ctx, minttypes.NewMinter(sdk.NewDec(1_000_000)))

	distributionParams := suite.app.DistrKeeper.GetParams(suite.ctx)
	distributionParams.BaseProposerReward = sdk.ZeroDec()
	distributionParams.BonusProposerReward = sdk.ZeroDec()
	distributionParams.CommunityTax = sdk.ZeroDec()
	suite.app.DistrKeeper.SetParams(suite.ctx, distributionParams)
}

func (suite *KeeperTestSuite) SetupDefaultPool() {
	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	poolId := suite.createGammPool([]string{bondDenom, "foo"})
	suite.Require().Equal(poolId, uint64(1))
}

// CreateRandomAccounts is a function return a list of randomly generated AccAddresses
func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

func (suite *KeeperTestSuite) createGammPool(denoms []string) uint64 {
	coins := suite.app.GAMMKeeper.GetParams(suite.ctx).PoolCreationFee
	poolAssets := []gammtypes.PoolAsset{}
	for _, denom := range denoms {
		coins = coins.Add(sdk.NewInt64Coin(denom, 1000000000000000000))
		poolAssets = append(poolAssets, gammtypes.PoolAsset{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(denom, sdk.NewInt(1000000000000000000)),
		})
	}

	acc1 := CreateRandomAccounts(1)[0]
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, acc1, coins)
	suite.Require().NoError(err)

	poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(
		suite.ctx, acc1, balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, poolAssets, "")
	suite.Require().NoError(err)

	return poolId
}

func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) (lockID uint64) {
	msgServer := lockupkeeper.NewMsgServerImpl(suite.app.LockupKeeper)
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
	suite.Require().NoError(err)
	msgResponse, err := msgServer.LockTokens(sdk.WrapSDKContext(suite.ctx), lockuptypes.NewMsgLockTokens(addr, duration, coins))
	suite.Require().NoError(err)
	return msgResponse.ID
}

func (suite *KeeperTestSuite) SetupValidators(bondStatuses []stakingtypes.BondStatus) []sdk.ValAddress {
	valAddrs := []sdk.ValAddress{}
	for _, status := range bondStatuses {
		valAddr := suite.helper.SetupValidator(status)
		valAddrs = append(valAddrs, valAddr)
	}
	return valAddrs
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
