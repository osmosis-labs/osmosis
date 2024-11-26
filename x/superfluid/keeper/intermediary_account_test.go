package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

func (s *KeeperTestSuite) TestIntermediaryAccountCreation() {
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		delegatorNumber  int64
		superDelegations []superfluidDelegation
	}{
		{
			"test intermediary account with single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
		},
		{
			"test multiple intermediary accounts with multiple superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
		},
		// Can create intermediary account with unbonded, unbonding validators
		{
			"test intermediary account with unbonded validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
		},
		{
			"test intermediary account with unbonding validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonding},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			valAddrs := s.SetupValidators(tc.validatorStats)
			delAddrs := CreateRandomAccounts(int(tc.delegatorNumber))

			// we create two additional pools: total three pools, 10 gauges
			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

			var interAccs []types.SuperfluidIntermediaryAccount

			for _, superDelegation := range tc.superDelegations {
				delAddr := delAddrs[superDelegation.delIndex]
				valAddr := valAddrs[superDelegation.valIndex]
				denom := denoms[superDelegation.lpIndex]

				// check intermediary Account prior to superfluid delegation, should have nil Intermediary Account
				expAcc := types.NewSuperfluidIntermediaryAccount(denom, valAddr.String(), 0)
				interAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, expAcc.GetAccAddress())
				s.Require().NotEqual(expAcc.GetAccAddress(), interAcc.GetAccAddress())
				s.Require().Equal("", interAcc.Denom)
				s.Require().Equal(uint64(0), interAcc.GaugeId)
				s.Require().Equal("", interAcc.ValAddr)

				lock := s.setupSuperfluidDelegate(delAddr, valAddr, denom, superDelegation.lpAmount)

				// check that intermediary Account connection is established
				interAccConnection := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lock.ID)
				s.Require().Equal(expAcc.GetAccAddress(), interAccConnection)

				interAcc = s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, interAccConnection)
				s.Require().Equal(expAcc.GetAccAddress(), interAcc.GetAccAddress())

				// check on interAcc that has been created
				s.Require().Equal(denom, interAcc.Denom)
				s.Require().Equal(valAddr.String(), interAcc.ValAddr)

				interAccs = append(interAccs, interAcc)
			}
			s.checkIntermediaryAccountDelegations(interAccs)
		})
	}
}

func (s *KeeperTestSuite) TestIntermediaryAccountsSetGetDeleteFlow() {
	s.SetupTest()

	// initial check
	accs := s.App.SuperfluidKeeper.GetAllIntermediaryAccounts(s.Ctx)
	s.Require().Len(accs, 0)

	// set account
	valAddr := sdk.ValAddress([]byte("addr1---------------"))
	acc := types.NewSuperfluidIntermediaryAccount(DefaultGammAsset, valAddr.String(), 1)
	s.App.SuperfluidKeeper.SetIntermediaryAccount(s.Ctx, acc)

	// get account
	gacc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, acc.GetAccAddress())
	s.Require().Equal(gacc.Denom, DefaultGammAsset)
	s.Require().Equal(gacc.ValAddr, valAddr.String())
	s.Require().Equal(gacc.GaugeId, uint64(1))

	// check accounts
	accs = s.App.SuperfluidKeeper.GetAllIntermediaryAccounts(s.Ctx)
	s.Require().Equal(accs, []types.SuperfluidIntermediaryAccount{acc})

	// delete asset
	s.App.SuperfluidKeeper.DeleteIntermediaryAccount(s.Ctx, acc.GetAccAddress())

	// get account
	gacc = s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, acc.GetAccAddress())
	s.Require().Equal(gacc.Denom, "")
	s.Require().Equal(gacc.ValAddr, "")
	s.Require().Equal(gacc.GaugeId, uint64(0))

	// check accounts
	accs = s.App.SuperfluidKeeper.GetAllIntermediaryAccounts(s.Ctx)
	s.Require().Len(accs, 0)
}

func (s *KeeperTestSuite) TestLockIdIntermediaryAccountConnection() {
	s.SetupTest()

	// get account
	addr := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, 1)
	s.Require().Equal(addr.String(), "")

	// set account
	valAddr := sdk.ValAddress([]byte("addr1---------------"))
	acc := types.NewSuperfluidIntermediaryAccount(DefaultGammAsset, valAddr.String(), 1)
	s.App.SuperfluidKeeper.SetLockIdIntermediaryAccountConnection(s.Ctx, 1, acc)

	// get account
	addr = s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, 1)
	s.Require().Equal(addr.String(), acc.GetAccAddress().String())

	// check get all
	conns := s.App.SuperfluidKeeper.GetAllLockIdIntermediaryAccountConnections(s.Ctx)
	s.Require().Len(conns, 1)

	// delete account
	s.App.SuperfluidKeeper.DeleteLockIdIntermediaryAccountConnection(s.Ctx, 1)

	// get account
	addr = s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, 1)
	s.Require().Equal(addr.String(), "")
}
