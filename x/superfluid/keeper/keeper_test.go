package keeper_test

import (
	"fmt"
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
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) GetSuite() *suite.Suite {
	return &suite.Suite
}
func (suite *KeeperTestSuite) GetCtx() sdk.Context {
	return suite.Ctx
}
func (suite *KeeperTestSuite) GetApp() *app.OsmosisApp {
	return suite.App
}
func (suite *KeeperTestSuite) SetCtx(Ctx sdk.Context) {
	suite.Ctx = Ctx
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.App = app.Setup(false)

	startTime := time.Unix(1645580000, 0)
	suite.Ctx = suite.App.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: startTime.UTC()})
	a := suite.App.StakingKeeper.GetParams(suite.Ctx).BondDenom
	fmt.Println(a)
	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.App.SuperfluidKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
	suite.SetupDefaultPool()

	unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

	suite.App.IncentivesKeeper.SetLockableDurations(suite.Ctx, []time.Duration{
		time.Hour * 24 * 14,
		time.Hour,
		time.Hour * 3,
		time.Hour * 7,
		unbondingDuration,
	})

	// TODO: Revisit if this is needed, it was added due to another bug in testing that is now fixed.
	epochIdentifier := suite.App.SuperfluidKeeper.GetEpochIdentifier(suite.Ctx)
	suite.App.EpochsKeeper.SetEpochInfo(suite.Ctx, epochtypes.EpochInfo{
		Identifier:              epochIdentifier,
		StartTime:               startTime,
		Duration:                time.Hour,
		CurrentEpochStartTime:   startTime,
		CurrentEpochStartHeight: 1,
		CurrentEpoch:            1,
		EpochCountingStarted:    true,
	})

	mintParams := suite.App.MintKeeper.GetParams(suite.Ctx)
	mintParams.EpochIdentifier = epochIdentifier
	mintParams.DistributionProportions = minttypes.DistributionProportions{
		Staking:          sdk.OneDec(),
		PoolIncentives:   sdk.ZeroDec(),
		DeveloperRewards: sdk.ZeroDec(),
		CommunityPool:    sdk.ZeroDec(),
	}
	suite.App.MintKeeper.SetParams(suite.Ctx, mintParams)
	suite.App.MintKeeper.SetMinter(suite.Ctx, minttypes.NewMinter(sdk.NewDec(1_000_000)))

	distributionParams := suite.App.DistrKeeper.GetParams(suite.Ctx)
	distributionParams.BaseProposerReward = sdk.ZeroDec()
	distributionParams.BonusProposerReward = sdk.ZeroDec()
	distributionParams.CommunityTax = sdk.ZeroDec()
	suite.App.DistrKeeper.SetParams(suite.Ctx, distributionParams)
}

func (suite *KeeperTestSuite) SetupDefaultPool() {
	bondDenom := suite.App.StakingKeeper.BondDenom(suite.Ctx)
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
	coins := suite.App.GAMMKeeper.GetParams(suite.Ctx).PoolCreationFee
	poolAssets := []gammtypes.PoolAsset{}
	for _, denom := range denoms {
		coins = coins.Add(sdk.NewInt64Coin(denom, 1000000000000000000))
		poolAssets = append(poolAssets, gammtypes.PoolAsset{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(denom, sdk.NewInt(1000000000000000000)),
		})
	}

	acc1 := CreateRandomAccounts(1)[0]
	err := simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, acc1, coins)
	suite.Require().NoError(err)

	poolId, err := suite.App.GAMMKeeper.CreateBalancerPool(
		suite.Ctx, acc1, balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, poolAssets, "")
	suite.Require().NoError(err)

	return poolId
}

func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) (lockID uint64) {
	msgServer := lockupkeeper.NewMsgServerImpl(suite.App.LockupKeeper)
	err := simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, coins)
	suite.Require().NoError(err)
	msgResponse, err := msgServer.LockTokens(sdk.WrapSDKContext(suite.Ctx), lockuptypes.NewMsgLockTokens(addr, duration, coins))
	suite.Require().NoError(err)
	return msgResponse.ID
}

func (suite *KeeperTestSuite) SetupValidators(bondStatuses []stakingtypes.BondStatus) []sdk.ValAddress {
	valAddrs := []sdk.ValAddress{}
	for _, status := range bondStatuses {
		valAddr := suite.SetupValidator(status)
		valAddrs = append(valAddrs, valAddr)
	}
	return valAddrs
}

func (suite *KeeperTestSuite) SetupGammPoolsAndSuperfluidAssets(multipliers []sdk.Dec) []string {
	suite.app.GAMMKeeper.SetParams(suite.ctx, gammtypes.Params{
		PoolCreationFee: sdk.Coins{},
	})

	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	denoms := []string{}

	for index, multiplier := range multipliers {
		token := fmt.Sprintf("token%d", index)

		params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
		uosmoAmount := gammtypes.InitPoolSharesSupply.ToDec().Mul(multiplier).Quo(sdk.OneDec().Sub(params.MinimumRiskFactor)).RoundInt()

		err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, acc1, sdk.NewCoins(
			sdk.NewCoin("uosmo", uosmoAmount.Mul(sdk.NewInt(10))),
			sdk.NewInt64Coin(token, 100000),
		))
		suite.NoError(err)

		var (
			defaultFutureGovernor = ""

			// pool assets
			defaultFooAsset gammtypes.PoolAsset = gammtypes.PoolAsset{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("uosmo", uosmoAmount),
			}
			defaultBarAsset gammtypes.PoolAsset = gammtypes.PoolAsset{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(token, sdk.NewInt(10000)),
			}
			poolAssets []gammtypes.PoolAsset = []gammtypes.PoolAsset{defaultFooAsset, defaultBarAsset}
		)

		poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(suite.ctx, acc1, balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, poolAssets, defaultFutureGovernor)
		suite.Require().NoError(err)

		pool, err := suite.app.GAMMKeeper.GetPool(suite.ctx, poolId)
		suite.Require().NoError(err)

		denom := pool.GetTotalShares().Denom
		suite.app.SuperfluidKeeper.AddNewSuperfluidAsset(suite.ctx, types.SuperfluidAsset{
			Denom:     denom,
			AssetType: types.SuperfluidAssetTypeLPShare,
		})

		denoms = append(denoms, denom)
	}
	return denoms
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
