package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v15/x/mint/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
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
	poolId, err := suite.App.PoolManagerKeeper.CreatePool(suite.Ctx, msg)
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

		err := suite.App.SuperfluidKeeper.AddNewSuperfluidAsset(suite.Ctx, types.SuperfluidAsset{
			Denom:     denom,
			AssetType: types.SuperfluidAssetTypeLPShare,
		})
		suite.Require().NoError(err)

		// register a LP token as a superfluid asset
		err = suite.App.SuperfluidKeeper.AddNewSuperfluidAsset(suite.Ctx, types.SuperfluidAsset{
			Denom:     denom,
			AssetType: types.SuperfluidAssetTypeLPShare,
		})
		suite.Require().NoError(err)

		denoms = append(denoms, denom)
		poolIds = append(poolIds, pool.GetId())
	}

	return denoms, poolIds
}

func (suite *KeeperTestSuite) setupSuperfluidDelegations(valAddrs []sdk.ValAddress, superDelegations []superfluidDelegation, denoms []string) ([]sdk.AccAddress, []types.SuperfluidIntermediaryAccount, []lockuptypes.PeriodLock) {
	flagIntermediaryAcc := make(map[string]bool)
	intermediaryAccs := []types.SuperfluidIntermediaryAccount{}
	locks := []lockuptypes.PeriodLock{}
	delAddrs := []sdk.AccAddress{}

	// we do sanity check on the test cases.
	// if superfluid staking is happening with single val and multiple superfluid delegations,
	// we should be running `AddTokensToLockByID`, instead of creating new locks
	delegatorAddressMap := make(map[int64]sdk.AccAddress)
	for _, superDelegation := range superDelegations {
		// either create or use existing delegator address
		var delAddr sdk.AccAddress
		if cachedAddr, ok := delegatorAddressMap[superDelegation.delIndex]; ok {
			delAddr = cachedAddr
		} else {
			pk := ed25519.GenPrivKey().PubKey()
			delAddr = sdk.AccAddress(pk.Address())
			delegatorAddressMap[superDelegation.delIndex] = delAddr
			delAddrs = append(delAddrs, delAddr)
		}

		valAddr := valAddrs[superDelegation.valIndex]
		lock := suite.setupSuperfluidDelegate(delAddr, valAddr, denoms[superDelegation.lpIndex], superDelegation.lpAmount)
		address := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lock.ID)
		gotAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, address)

		// save accounts for future use
		if flagIntermediaryAcc[gotAcc.String()] == false {
			flagIntermediaryAcc[gotAcc.String()] = true
			intermediaryAccs = append(intermediaryAccs, gotAcc)
		}
		// save locks for future use
		locks = append(locks, lock)
	}
	return delAddrs, intermediaryAccs, locks
}

func (suite *KeeperTestSuite) checkIntermediaryAccountDelegations(intermediaryAccs []types.SuperfluidIntermediaryAccount) {
	for _, acc := range intermediaryAccs {
		valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
		suite.Require().NoError(err)

		// check delegation from intermediary account to validator
		delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, acc.GetAccAddress(), valAddr)
		suite.Require().True(found)
		suite.Require().True(delegation.Shares.GTE(sdk.NewDec(10000000)))

		// check delegated tokens
		validator, found := suite.App.StakingKeeper.GetValidator(suite.Ctx, valAddr)
		suite.Require().True(found)
		delegatedTokens := validator.TokensFromShares(delegation.Shares).TruncateInt()
		suite.Require().True(delegatedTokens.GTE(sdk.NewInt(10000000)))
	}
}

func (suite *KeeperTestSuite) setupSuperfluidDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, denom string, amount int64) lockuptypes.PeriodLock {
	unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

	// create lockup of LP token
	coins := sdk.Coins{sdk.NewInt64Coin(denom, amount)}
	lastLockID := suite.App.LockupKeeper.GetLastLockID(suite.Ctx)

	lockID := suite.LockTokens(delAddr, coins, unbondingDuration)
	lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockID)
	suite.Require().NoError(err)

	// here we check if check `LockTokens` added to existing locks or created a new lock.
	// if `LockTokens` created a new lock, we continue SuperfluidDelegate
	// if lock has been existing before, we wouldn't have to call SuperfluidDelegate separately, as hooks on LockTokens would have automatically called IncreaseSuperfluidDelegation
	if lastLockID != lockID {
		err = suite.App.SuperfluidKeeper.SuperfluidDelegate(suite.Ctx, lock.Owner, lock.ID, valAddr.String())
		suite.Require().NoError(err)
	} else {
		// here we handle two cases.
		// 1. the lock has existed before but has not been superflud staking
		// 2. the lock has existed before and has been superfluid staking

		// we check if synth lock that has existed before, if it did, it means that the lock has been superfluid staked
		// we do not care about unbonding synthlocks, as superfluid delegation has no effect

		_, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
		// if lock has existed before but has not been superfluid staked, we do initial superfluid staking
		if err != nil {
			err = suite.App.SuperfluidKeeper.SuperfluidDelegate(suite.Ctx, lock.Owner, lock.ID, valAddr.String())
			suite.Require().NoError(err)
		}
	}

	return *lock
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
