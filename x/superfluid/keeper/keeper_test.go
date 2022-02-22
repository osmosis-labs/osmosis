package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	queryClient types.QueryClient
	app         *app.OsmosisApp
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})

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
}

func (suite *KeeperTestSuite) SetupDefaultPool() {
	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	poolId := suite.createGammPool([]string{bondDenom, "foo"})
	suite.Require().Equal(poolId, uint64(1))
}

func (suite *KeeperTestSuite) BeginNewBlock(executeNextEpoch bool) {
	epochIdentifier := suite.app.SuperfluidKeeper.GetParams(suite.ctx).RefreshEpochIdentifier
	epoch := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, epochIdentifier)
	newBlockTime := suite.ctx.BlockTime().Add(5 * time.Second)
	if executeNextEpoch {
		endEpochTime := epoch.CurrentEpochStartTime.Add(epoch.Duration)
		newBlockTime = endEpochTime.Add(time.Second)
	}
	header := tmproto.Header{Height: suite.ctx.BlockHeight() + 1, Time: newBlockTime}
	reqBeginBlock := abci.RequestBeginBlock{Header: header}
	suite.app.BeginBlocker(suite.ctx, reqBeginBlock)

}

func (suite *KeeperTestSuite) EndBlock() {
	reqEndBlock := abci.RequestEndBlock{Height: suite.ctx.BlockHeight()}
	suite.app.EndBlocker(suite.ctx, reqEndBlock)
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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
