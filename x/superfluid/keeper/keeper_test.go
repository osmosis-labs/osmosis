package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	epochtypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	queryClient types.QueryClient
	app         *app.OsmosisApp
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)

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
}

func (suite *KeeperTestSuite) SetupDefaultPool() {
	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	poolId := suite.createGammPool([]string{bondDenom, "foo"})
	suite.Require().Equal(poolId, uint64(1))
}

func (suite *KeeperTestSuite) BeginNewBlock(executeNextEpoch bool) {
	epochIdentifier := suite.app.SuperfluidKeeper.GetEpochIdentifier(suite.ctx)
	epoch := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, epochIdentifier)
	newBlockTime := suite.ctx.BlockTime().Add(5 * time.Second)
	if executeNextEpoch {
		endEpochTime := epoch.CurrentEpochStartTime.Add(epoch.Duration)
		newBlockTime = endEpochTime.Add(time.Second)
	}
	// fmt.Println(executeNextEpoch, suite.ctx.BlockTime(), newBlockTime)
	header := tmproto.Header{Height: suite.ctx.BlockHeight() + 1, Time: newBlockTime}
	suite.ctx = suite.ctx.WithBlockTime(newBlockTime).WithBlockHeight(suite.ctx.BlockHeight() + 1)
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

func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) (lockID uint64) {
	msgServer := lockupkeeper.NewMsgServerImpl(*suite.app.LockupKeeper)
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
	suite.Require().NoError(err)
	msgResponse, err := msgServer.LockTokens(sdk.WrapSDKContext(suite.ctx), lockuptypes.NewMsgLockTokens(addr, duration, coins))
	suite.Require().NoError(err)
	return msgResponse.ID
}

func (suite *KeeperTestSuite) SetupValidator(bondStatus stakingtypes.BondStatus) sdk.ValAddress {
	valPub := secp256k1.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(valPub.Address())
	bondDenom := suite.app.StakingKeeper.GetParams(suite.ctx).BondDenom
	selfBond := sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(100), Denom: bondDenom})

	simapp.FundAccount(suite.app.BankKeeper, suite.ctx, sdk.AccAddress(valAddr), selfBond)
	sh := teststaking.NewHelper(suite.T(), suite.ctx, *suite.app.StakingKeeper)
	msg := sh.CreateValidatorMsg(valAddr, valPub, selfBond[0].Amount)
	sh.Handle(msg, true)
	val, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
	suite.Require().True(found)
	val = val.UpdateStatus(bondStatus)
	suite.app.StakingKeeper.SetValidator(suite.ctx, val)

	return valAddr
}

func (suite *KeeperTestSuite) SetupValidators(bondStatuses []stakingtypes.BondStatus) []sdk.ValAddress {
	valAddrs := []sdk.ValAddress{}
	for _, status := range bondStatuses {
		valAddr := suite.SetupValidator(status)
		valAddrs = append(valAddrs, valAddr)
	}
	return valAddrs
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
