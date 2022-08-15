package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v11/app/apptesting"
	epochtypes "github.com/osmosis-labs/osmosis/v11/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v11/x/gamm/types"
	minttypes "github.com/osmosis-labs/osmosis/v11/x/mint/types"
	"github.com/osmosis-labs/osmosis/v11/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v11/x/superfluid/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
	querier     keeper.Querier
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
	suite.querier = keeper.NewQuerier(*suite.App.SuperfluidKeeper)

	startTime := suite.Ctx.BlockHeader().Time

	unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

	suite.App.IncentivesKeeper.SetLockableDurations(suite.Ctx, []time.Duration{
		time.Hour * 24 * 14,
		time.Hour,
		time.Hour * 3,
		time.Hour * 7,
		unbondingDuration,
	})

	superfluidEpochIdentifer := "superfluid_epoch"
	incentiveKeeperParams := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
	incentiveKeeperParams.DistrEpochIdentifier = superfluidEpochIdentifer
	suite.App.IncentivesKeeper.SetParams(suite.Ctx, incentiveKeeperParams)
	err := suite.App.EpochsKeeper.AddEpochInfo(suite.Ctx, epochtypes.EpochInfo{
		Identifier:              superfluidEpochIdentifer,
		StartTime:               startTime,
		Duration:                time.Hour,
		CurrentEpochStartTime:   startTime,
		CurrentEpochStartHeight: 1,
		CurrentEpoch:            1,
		EpochCountingStarted:    true,
	})
	suite.Require().NoError(err)

	mintParams := suite.App.MintKeeper.GetParams(suite.Ctx)
	mintParams.EpochIdentifier = superfluidEpochIdentifer
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
	poolAssets := []balancer.PoolAsset{}
	for _, denom := range denoms {
		coins = coins.Add(sdk.NewInt64Coin(denom, 1000000000000000000))
		poolAssets = append(poolAssets, balancer.PoolAsset{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(denom, sdk.NewInt(1000000000000000000)),
		})
	}

	acc1 := CreateRandomAccounts(1)[0]
	suite.FundAcc(acc1, coins)

	msg := balancer.NewMsgCreateBalancerPool(acc1, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}, poolAssets, "")
	poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
	suite.Require().NoError(err)

	return poolId
}

func (suite *KeeperTestSuite) SetupValidators(bondStatuses []stakingtypes.BondStatus) []sdk.ValAddress {
	valAddrs := []sdk.ValAddress{}
	for _, status := range bondStatuses {
		valAddr := suite.SetupValidator(status)
		valAddrs = append(valAddrs, valAddr)
	}
	return valAddrs
}

func (suite *KeeperTestSuite) SetupGammPoolsAndSuperfluidAssets(multipliers []sdk.Dec) ([]string, []uint64) {
	pools := suite.SetupGammPoolsWithBondDenomMultiplier(multipliers)

	denoms := []string{}
	poolIds := []uint64{}
	for _, pool := range pools {
		denom := gammtypes.GetPoolShareDenom(pool.GetId())

		suite.App.SuperfluidKeeper.AddNewSuperfluidAsset(suite.Ctx, types.SuperfluidAsset{
			Denom:     denom,
			AssetType: types.SuperfluidAssetTypeLPShare,
		})

		// register a LP token as a superfluid asset
		suite.App.SuperfluidKeeper.AddNewSuperfluidAsset(suite.Ctx, types.SuperfluidAsset{
			Denom:     denom,
			AssetType: types.SuperfluidAssetTypeLPShare,
		})

		denoms = append(denoms, denom)
		poolIds = append(poolIds, pool.GetId())
	}

	return denoms, poolIds
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
