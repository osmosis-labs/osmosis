package keeper_test

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	incentivetypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

var DefaultGammAsset = "gamm/pool/1"

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
	querier     keeper.Querier
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.querier = keeper.NewQuerier(*s.App.SuperfluidKeeper)

	startTime := s.Ctx.BlockHeader().Time

	stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	unbondingDuration := stakingParams.UnbondingTime

	s.App.IncentivesKeeper.SetLockableDurations(s.Ctx, []time.Duration{
		time.Hour * 24 * 14,
		time.Hour,
		time.Hour * 3,
		time.Hour * 7,
		unbondingDuration,
	})

	superfluidEpochIdentifer := "superfluid_epoch"
	incentiveKeeperParams := s.App.IncentivesKeeper.GetParams(s.Ctx)
	incentiveKeeperParams.DistrEpochIdentifier = superfluidEpochIdentifer
	s.App.IncentivesKeeper.SetParams(s.Ctx, incentiveKeeperParams)
	err = s.App.EpochsKeeper.AddEpochInfo(s.Ctx, epochtypes.EpochInfo{
		Identifier:              superfluidEpochIdentifer,
		StartTime:               startTime,
		Duration:                time.Hour,
		CurrentEpochStartTime:   startTime,
		CurrentEpochStartHeight: 1,
		CurrentEpoch:            1,
		EpochCountingStarted:    true,
	})
	s.Require().NoError(err)

	mintParams := s.App.MintKeeper.GetParams(s.Ctx)
	mintParams.EpochIdentifier = superfluidEpochIdentifer
	mintParams.DistributionProportions = minttypes.DistributionProportions{
		Staking:          osmomath.OneDec(),
		PoolIncentives:   osmomath.ZeroDec(),
		DeveloperRewards: osmomath.ZeroDec(),
		CommunityPool:    osmomath.ZeroDec(),
	}
	s.App.MintKeeper.SetParams(s.Ctx, mintParams)
	s.App.MintKeeper.SetMinter(s.Ctx, minttypes.NewMinter(osmomath.NewDec(1_000_000)))

	distributionParams, err := s.App.DistrKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	distributionParams.BaseProposerReward = osmomath.ZeroDec()
	distributionParams.BonusProposerReward = osmomath.ZeroDec()
	distributionParams.CommunityTax = osmomath.ZeroDec()
	s.App.DistrKeeper.Params.Set(s.Ctx, distributionParams)
	s.App.IncentivesKeeper.SetParam(s.Ctx, incentivetypes.KeyMinValueForDistr, sdk.NewCoin("stake", osmomath.NewInt(1)))
}

func (s *KeeperTestSuite) SetupDefaultPool() {
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	poolId := s.createGammPool([]string{bondDenom, "foo"})
	s.Require().Equal(poolId, uint64(1))
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

func (s *KeeperTestSuite) createGammPool(denoms []string) uint64 {
	coins := s.App.GAMMKeeper.GetParams(s.Ctx).PoolCreationFee
	poolAssets := []balancer.PoolAsset{}
	for _, denom := range denoms {
		coins = coins.Add(sdk.NewInt64Coin(denom, 1000000000000000000))
		poolAssets = append(poolAssets, balancer.PoolAsset{
			Weight: osmomath.NewInt(100),
			Token:  sdk.NewCoin(denom, osmomath.NewInt(1000000000000000000)),
		})
	}

	acc1 := CreateRandomAccounts(1)[0]
	s.FundAcc(acc1, coins)

	msg := balancer.NewMsgCreateBalancerPool(acc1, balancer.PoolParams{
		SwapFee: osmomath.NewDecWithPrec(1, 2),
		ExitFee: osmomath.ZeroDec(),
	}, poolAssets, "")
	poolId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
	s.Require().NoError(err)

	return poolId
}

func (s *KeeperTestSuite) SetupValidators(bondStatuses []stakingtypes.BondStatus) []sdk.ValAddress {
	valAddrs := []sdk.ValAddress{}
	for _, status := range bondStatuses {
		valAddr := s.SetupValidator(status)
		valAddrs = append(valAddrs, valAddr)
	}
	return valAddrs
}

func (s *KeeperTestSuite) SetupGammPoolsAndSuperfluidAssets(multipliers []osmomath.Dec) ([]string, []uint64) {
	pools := s.SetupGammPoolsWithBondDenomMultiplier(multipliers)

	denoms := []string{}
	poolIds := []uint64{}
	for _, pool := range pools {
		denom := gammtypes.GetPoolShareDenom(pool.GetId())

		err := s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
			Denom:     denom,
			AssetType: types.SuperfluidAssetTypeLPShare,
		})
		s.Require().NoError(err)

		// register a LP token as a superfluid asset
		err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
			Denom:     denom,
			AssetType: types.SuperfluidAssetTypeLPShare,
		})
		s.Require().NoError(err)

		denoms = append(denoms, denom)
		poolIds = append(poolIds, pool.GetId())
	}

	return denoms, poolIds
}

func (s *KeeperTestSuite) setupSuperfluidDelegations(valAddrs []sdk.ValAddress, superDelegations []superfluidDelegation, denoms []string) ([]sdk.AccAddress, []types.SuperfluidIntermediaryAccount, []lockuptypes.PeriodLock) {
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
		lock := s.setupSuperfluidDelegate(delAddr, valAddr, denoms[superDelegation.lpIndex], superDelegation.lpAmount)
		address := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lock.ID)
		gotAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, address)

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

func (s *KeeperTestSuite) checkIntermediaryAccountDelegations(intermediaryAccs []types.SuperfluidIntermediaryAccount) {
	for _, acc := range intermediaryAccs {
		valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
		s.Require().NoError(err)

		// check delegation from intermediary account to validator
		delegation, err := s.App.StakingKeeper.GetDelegation(s.Ctx, acc.GetAccAddress(), valAddr)
		s.Require().NoError(err)
		s.Require().True(delegation.Shares.GTE(osmomath.NewDec(10000000)))

		// check delegated tokens
		validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
		s.Require().NoError(err)
		delegatedTokens := validator.TokensFromShares(delegation.Shares).TruncateInt()
		s.Require().True(delegatedTokens.GTE(osmomath.NewInt(10000000)))
	}
}

func (s *KeeperTestSuite) setupSuperfluidDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, denom string, amount int64) lockuptypes.PeriodLock {
	stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	unbondingDuration := stakingParams.UnbondingTime

	// create lockup of LP token
	coins := sdk.Coins{sdk.NewInt64Coin(denom, amount)}
	lastLockID := s.App.LockupKeeper.GetLastLockID(s.Ctx)

	lockID := s.LockTokens(delAddr, coins, unbondingDuration)
	lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockID)
	s.Require().NoError(err)

	// here we check if check `LockTokens` added to existing locks or created a new lock.
	// if `LockTokens` created a new lock, we continue SuperfluidDelegate
	// if lock has been existing before, we wouldn't have to call SuperfluidDelegate separately, as hooks on LockTokens would have automatically called IncreaseSuperfluidDelegation
	if lastLockID != lockID {
		err = s.App.SuperfluidKeeper.SuperfluidDelegate(s.Ctx, lock.Owner, lock.ID, valAddr.String())
		s.Require().NoError(err)
	} else {
		// here we handle two cases.
		// 1. the lock has existed before but has not been superflud staking
		// 2. the lock has existed before and has been superfluid staking

		// we check if synth lock that has existed before, if it did, it means that the lock has been superfluid staked
		// we do not care about unbonding synthlocks, as superfluid delegation has no effect

		_, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lockID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
		// if lock has existed before but has not been superfluid staked, we do initial superfluid staking
		if err != nil {
			err = s.App.SuperfluidKeeper.SuperfluidDelegate(s.Ctx, lock.Owner, lock.ID, valAddr.String())
			s.Require().NoError(err)
		}
	}

	return *lock
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
