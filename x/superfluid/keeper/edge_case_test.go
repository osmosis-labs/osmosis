package keeper_test

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/x/evidence/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

func (s *KeeperTestSuite) TestSuperfluidDelegatedValidatorJailed() {
	testCases := []struct {
		name                 string
		validatorStats       []stakingtypes.BondStatus
		delegatorNumber      int
		superDelegations     []superfluidDelegation
		jailedValIndexes     []int64
		expJailedLockIndexes []int64
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]int64{0},
			[]int64{0},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

			locks := []lockuptypes.PeriodLock{}
			slashFactor, err := s.App.SlashingKeeper.SlashFractionDoubleSign(s.Ctx)
			s.Require().NoError(err)

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				delAddr := delAddrs[del.delIndex]
				lock := s.setupSuperfluidDelegate(delAddr, valAddr, denoms[del.lpIndex], del.lpAmount)

				// save accounts and locks for future use
				locks = append(locks, lock)
			}

			// slash validator
			for _, valIndex := range tc.jailedValIndexes {
				validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddrs[valIndex])
				s.Require().NoError(err)
				s.Ctx = s.Ctx.WithBlockHeight(100)
				consAddr, err := validator.GetConsAddr()
				s.Require().NoError(err)
				// slash by slash factor
				power := sdk.TokensToConsensusPower(validator.Tokens, sdk.DefaultPowerReduction)

				// Note: this calls BeforeValidatorSlashed hook
				s.handleEquivocationEvidence(s.Ctx, &evidencetypes.Equivocation{
					Height:           80,
					Time:             time.Time{},
					Power:            power,
					ConsensusAddress: sdk.ConsAddress(consAddr).String(),
				})
				val, err := s.App.StakingKeeper.GetValidatorByConsAddr(s.Ctx, consAddr)
				s.Require().NoError(err)
				s.Require().Equal(val.Jailed, true)
			}

			// check lock changes after validator & lockups slashing
			for _, lockIndex := range tc.expJailedLockIndexes {
				gotLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, locks[lockIndex].ID)
				s.Require().NoError(err)
				s.Require().Equal(
					gotLock.Coins.AmountOf(denoms[0]).String(),
					osmomath.NewDec(1000000).Mul(osmomath.OneDec().Sub(slashFactor)).TruncateInt().String(),
				)
			}
		})
	}
}

func (s *KeeperTestSuite) TestTryUnbondingSuperfluidLockupDirectly() {
	testCases := []struct {
		name               string
		validatorStats     []stakingtypes.BondStatus
		delegatorNumber    int
		superDelegations   []superfluidDelegation
		expInterDelegation []osmomath.Dec
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]osmomath.Dec{osmomath.NewDec(10000000)}, // 50% x 20 x 1000000
		},
		{
			"with single validator and additional superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]osmomath.Dec{osmomath.NewDec(20000000)}, // 50% x 20 x 1000000 x 2
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]osmomath.Dec{osmomath.NewDec(10000000), osmomath.NewDec(10000000)}, // 50% x 20 x 1000000
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)
			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

			// setup superfluid delegations
			_, _, locks := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)

			for _, lock := range locks {
				_, err := s.App.LockupKeeper.BeginUnlock(s.Ctx, lock.ID, sdk.Coins{})
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) handleEquivocationEvidence(ctx context.Context, evidence *types.Equivocation) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	consAddr := evidence.GetConsensusAddress(s.App.StakingKeeper.ConsensusAddressCodec())

	validator, err := s.App.StakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		return err
	}
	if validator == nil || validator.IsUnbonded() {
		return nil
	}

	if len(validator.GetOperator()) != 0 {
		if _, err := s.App.SlashingKeeper.GetPubkey(ctx, consAddr.Bytes()); err != nil {
			return nil
		}
	}

	// calculate the age of the evidence
	infractionHeight := evidence.GetHeight()
	infractionTime := evidence.GetTime()
	ageDuration := sdkCtx.BlockHeader().Time.Sub(infractionTime)
	ageBlocks := sdkCtx.BlockHeader().Height - infractionHeight

	// Reject evidence if the double-sign is too old. Evidence is considered stale
	// if the difference in time and number of blocks is greater than the allowed
	// parameters defined.
	cp := sdkCtx.ConsensusParams()
	if cp.Evidence != nil {
		if ageDuration > cp.Evidence.MaxAgeDuration && ageBlocks > cp.Evidence.MaxAgeNumBlocks {
			return nil
		}
	}

	if ok := s.App.SlashingKeeper.HasValidatorSigningInfo(ctx, consAddr); !ok {
		panic(fmt.Sprintf("expected signing info for validator %s but not found", consAddr))
	}

	// ignore if the validator is already tombstoned
	if s.App.SlashingKeeper.IsTombstoned(ctx, consAddr) {
		return nil
	}

	distributionHeight := infractionHeight - sdk.ValidatorUpdateDelay

	slashFractionDoubleSign, err := s.App.SlashingKeeper.SlashFractionDoubleSign(ctx)
	if err != nil {
		return err
	}

	err = s.App.SlashingKeeper.SlashWithInfractionReason(
		ctx,
		consAddr,
		slashFractionDoubleSign,
		evidence.GetValidatorPower(), distributionHeight,
		stakingtypes.Infraction_INFRACTION_DOUBLE_SIGN,
	)
	if err != nil {
		return err
	}

	// Jail the validator if not already jailed. This will begin unbonding the
	// validator if not already unbonding (tombstoned).
	if !validator.IsJailed() {
		err = s.App.SlashingKeeper.Jail(ctx, consAddr)
		if err != nil {
			return err
		}
	}

	err = s.App.SlashingKeeper.JailUntil(ctx, consAddr, types.DoubleSignJailEndTime)
	if err != nil {
		return err
	}

	err = s.App.SlashingKeeper.Tombstone(ctx, consAddr)
	if err != nil {
		return err
	}
	return s.App.EvidenceKeeper.Evidences.Set(ctx, evidence.Hash(), evidence)
}
