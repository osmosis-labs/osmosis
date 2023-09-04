package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
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
			slashFactor := s.App.SlashingKeeper.SlashFractionDoubleSign(s.Ctx)

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
				validator, found := s.App.StakingKeeper.GetValidator(s.Ctx, valAddrs[valIndex])
				s.Require().True(found)
				s.Ctx = s.Ctx.WithBlockHeight(100)
				consAddr, err := validator.GetConsAddr()
				s.Require().NoError(err)
				// slash by slash factor
				power := sdk.TokensToConsensusPower(validator.Tokens, sdk.DefaultPowerReduction)

				// Note: this calls BeforeValidatorSlashed hook
				s.App.EvidenceKeeper.HandleEquivocationEvidence(s.Ctx, &evidencetypes.Equivocation{
					Height:           80,
					Time:             time.Time{},
					Power:            power,
					ConsensusAddress: consAddr.String(),
				})
				val, found := s.App.StakingKeeper.GetValidatorByConsAddr(s.Ctx, consAddr)
				s.Require().True(found)
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
