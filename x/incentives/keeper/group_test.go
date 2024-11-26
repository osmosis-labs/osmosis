package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v27/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	poolincentivetypes "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

type createGroupTestCase struct {
	name             string
	coins            sdk.Coins
	numEpochPaidOver uint64
	// 0 by default unless overwritten
	creatorAddressIndex int
	poolIDs             []uint64
	// corresponds to the pool IDs above
	poolVolumesToSet []osmomath.Int

	expectedGaugeInfo           types.InternalGaugeInfo
	expectedPerpeutalGroupGauge bool
	expectErr                   error
}

// index of s.TestAccs that gets funded
const (
	// for every test case, receives the group creation fee and gauge tokens
	// ensuring it always has enough funds.
	fullyFundedAddressIndex = 0
	// has enough funds to pay group creation fee once and nothing else
	oneTimeFeeFundedIndex = 1
	// does not get funded with anything
	noFundingIndex = 2
)

var (
	makeDefaultSuccessCreateGroupTestCases = func(poolInfo apptesting.SupportedPoolAndGaugeInfo, concentratedGaugeRecord, balancerGaugeRecord, stableSwapGaugeRecord types.InternalGaugeRecord) []createGroupTestCase {
		return []createGroupTestCase{
			{
				name:             "two pools - created perpetual group gauge",
				coins:            defaultCoins,
				numEpochPaidOver: types.PerpetualNumEpochsPaidOver,
				poolIDs:          []uint64{poolInfo.ConcentratedPoolID, poolInfo.BalancerPoolID},
				poolVolumesToSet: []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount.Add(defaultVolumeAmount)},

				expectedPerpeutalGroupGauge: true,
				expectedGaugeInfo: addGaugeRecords(defaultEmptyGaugeInfo, []types.InternalGaugeRecord{
					concentratedGaugeRecord,
					balancerGaugeRecord,
				}),
			},
			{
				name:             "all incentive supported pools - created perpetual group gauge",
				coins:            defaultCoins,
				numEpochPaidOver: types.PerpetualNumEpochsPaidOver,
				poolIDs:          []uint64{poolInfo.ConcentratedPoolID, poolInfo.BalancerPoolID, poolInfo.StableSwapPoolID},
				poolVolumesToSet: []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount, defaultVolumeAmount},

				expectedPerpeutalGroupGauge: true,
				expectedGaugeInfo: addGaugeRecords(defaultEmptyGaugeInfo, []types.InternalGaugeRecord{
					concentratedGaugeRecord,
					balancerGaugeRecord,
					stableSwapGaugeRecord,
				}),
			},
			{
				name:             "two pools - created non-perpetual group gauge",
				coins:            defaultCoins,
				numEpochPaidOver: types.PerpetualNumEpochsPaidOver + 1,
				poolIDs:          []uint64{poolInfo.ConcentratedPoolID, poolInfo.BalancerPoolID},
				poolVolumesToSet: []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount},

				expectedPerpeutalGroupGauge: false, // explicit for clarity
				expectedGaugeInfo: addGaugeRecords(defaultEmptyGaugeInfo, []types.InternalGaugeRecord{
					concentratedGaugeRecord,
					balancerGaugeRecord,
				}),
			},

			{
				name:             "all incentive supported pools with custom amount - created non-perpetual group gauge",
				coins:            defaultCoins.Add(defaultCoins...).Add(defaultCoins...),
				numEpochPaidOver: types.PerpetualNumEpochsPaidOver + 4,
				poolIDs:          []uint64{poolInfo.ConcentratedPoolID, poolInfo.BalancerPoolID, poolInfo.StableSwapPoolID},
				poolVolumesToSet: []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount, defaultVolumeAmount},

				expectedPerpeutalGroupGauge: false, // explicit for clarity
				expectedGaugeInfo: addGaugeRecords(defaultEmptyGaugeInfo, []types.InternalGaugeRecord{
					concentratedGaugeRecord,
					balancerGaugeRecord,
					stableSwapGaugeRecord,
				}),
			},
		}
	}

	makeDefaultErrorCases = func(poolInfo apptesting.SupportedPoolAndGaugeInfo) []createGroupTestCase {
		return []createGroupTestCase{
			{
				name:             "error: fails to initialize group gauge due to cosmwasm pool that does not support incentives",
				coins:            defaultCoins,
				numEpochPaidOver: types.PerpetualNumEpochsPaidOver,
				poolIDs:          []uint64{poolInfo.BalancerPoolID, poolInfo.CosmWasmPoolID},
				poolVolumesToSet: []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount},

				expectErr: poolincentivetypes.UnsupportedPoolTypeError{PoolID: poolInfo.CosmWasmPoolID, PoolType: poolmanagertypes.CosmWasm},
			},

			{
				name:                "error: owner does not have enough funds to create gauge but has the fee",
				coins:               defaultCoins,
				creatorAddressIndex: oneTimeFeeFundedIndex,
				numEpochPaidOver:    types.PerpetualNumEpochsPaidOver,
				poolIDs:             []uint64{poolInfo.BalancerPoolID, poolInfo.ConcentratedPoolID},
				poolVolumesToSet:    []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount},
				expectErr:           fmt.Errorf("spendable balance 0%s is smaller than %s: insufficient funds", appparams.BaseCoinUnit, defaultCoins),
			},
			{
				name:                "error: owner does not have enough funds to pay creation fee",
				coins:               defaultCoins,
				creatorAddressIndex: noFundingIndex,
				numEpochPaidOver:    types.PerpetualNumEpochsPaidOver,
				poolIDs:             []uint64{poolInfo.BalancerPoolID, poolInfo.ConcentratedPoolID},
				poolVolumesToSet:    []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount},
				expectErr:           fmt.Errorf("spendable balance 0%s is smaller than %s: insufficient funds", feeDenom, customGroupCreationFee),
			},
			{
				name:             "error: duplicate pool IDs",
				coins:            defaultCoins,
				numEpochPaidOver: types.PerpetualNumEpochsPaidOver,
				poolIDs:          []uint64{poolInfo.ConcentratedPoolID, poolInfo.BalancerPoolID, poolInfo.ConcentratedPoolID},
				poolVolumesToSet: []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount.Add(defaultVolumeAmount)},
				expectErr:        types.DuplicatePoolIDError{PoolIDs: []uint64{poolInfo.ConcentratedPoolID, poolInfo.BalancerPoolID, poolInfo.ConcentratedPoolID}},
			},
		}
	}
)

// We have 3 different implementations of CreateGroup:
// - CreateGroup - standard client-facing
// - CreateGroupAsIncentivesModuleAcc - only callable by incentives module account
// - CreateGroupInternal - internal, used by the other two
// This test validates each of these implementations.
// While they are similar in functionality, they still differ.
// As a result, we create this shared test case to avoid code duplication but
// direct the test cases to the appropriate test function for custom logic.
func (s *KeeperTestSuite) TestCreateGroup() {

	type systemUnderTestType int
	const (
		CreateGroup systemUnderTestType = iota
		CreateGroupInternal
		CreateGroupAsIncentivesModuleAcc
		TotalTypes
	)

	differentExecutionTypeCount := 0

	for _, tc := range []struct {
		name                string
		systemUnderTestType systemUnderTestType
	}{
		{
			// Validates that the Group is created as defined by the CreateGroup spec with the
			// associated 1:1 group Gauge and the correct gauge records relating to the given pools'
			// internal perpetual gauge IDs.
			//
			// The test structure is that a general shared state setup is performed once at the top.
			// For every test case, we fund the same account with appropriate amounts, ensuring
			// that this account has a sufficient balance to pay for the group creation fee and transfer the
			// gauge tokens.
			//
			// For testing low balance error cases, we operate on other accounts that may or may not have
			// enough funds to pay for the group creation fee.
			name:                "CreateGroup",
			systemUnderTestType: CreateGroup,
		},
		{
			// This test:
			// Performs the same validations as TestCreateGroup with the exceptions:
			// - Group is not written to state (validation performed on the Group returned by CreateGroupInternal)
			// - Volume is not attempted to be synched, as a result, the initial weights are all zero.
			name:                "CreateGroupInternal",
			systemUnderTestType: CreateGroupInternal,
		},
		{
			// This test:
			// Performs the same validations as TestCreateGroupInternal with the exceptions:
			// - Group is written to start
			// - Can only be run by an incentives module account
			// - syncing volume is not attempted (initialized to zero)
			name:                "CreateGroupAsIncentivesModuleAcc",
			systemUnderTestType: CreateGroupAsIncentivesModuleAcc,
		},
	} {
		tc := tc
		s.Run(tc.name, func() {
			// We setup test state once and reuse it for all test cases
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultTime)

			// Create 4 pools of each possible type
			poolInfo := s.PrepareAllSupportedPools()

			expectedGroupGaugeId := s.App.IncentivesKeeper.GetLastGaugeID(s.Ctx) + 1

			// Initialize expected gauge records
			var (
				concentratedGaugeRecord = withRecordGaugeId(defaultZeroWeightGaugeRecord, poolInfo.ConcentratedGaugeID)
				balancerGaugeRecord     = withRecordGaugeId(defaultZeroWeightGaugeRecord, poolInfo.BalancerGaugeID)
				stableSwapGaugeRecord   = withRecordGaugeId(defaultZeroWeightGaugeRecord, poolInfo.StableSwapGaugeID)
			)

			// Set a custom creation fee to avoid test balances having false positives
			// due to having OSMO added during test setup
			s.App.IncentivesKeeper.SetParam(s.Ctx, types.KeyGroupCreationFee, customGroupCreationFee)

			// Fund fee once to a specific test account
			s.FundAcc(s.TestAccs[oneTimeFeeFundedIndex], customGroupCreationFee)

			communityPoolAddress := s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
			originalCommunityPoolBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)
			expectedCommunityPoolFeeBalance := sdk.NewCoins()

			if tc.systemUnderTestType == CreateGroup {
				s.runCreateGroupTests(poolInfo, concentratedGaugeRecord, balancerGaugeRecord, stableSwapGaugeRecord, expectedCommunityPoolFeeBalance, expectedGroupGaugeId, originalCommunityPoolBalance)
				differentExecutionTypeCount++
			} else if tc.systemUnderTestType == CreateGroupInternal {
				s.runCreateGroupInternalTests(poolInfo, concentratedGaugeRecord, balancerGaugeRecord, stableSwapGaugeRecord, expectedCommunityPoolFeeBalance, expectedGroupGaugeId, originalCommunityPoolBalance)
				differentExecutionTypeCount++
			} else if tc.systemUnderTestType == CreateGroupAsIncentivesModuleAcc {
				s.runCreateGroupAsIncentivesModuleAccTests(poolInfo, concentratedGaugeRecord, balancerGaugeRecord, expectedCommunityPoolFeeBalance, expectedGroupGaugeId, originalCommunityPoolBalance)
				differentExecutionTypeCount++
			} else {
				s.FailNow("unknown system under test type")
			}
		})
	}
	// Ensure that we ran all the tests to prevent false positions
	s.Require().Equal(int(TotalTypes), differentExecutionTypeCount)
}

func (s *KeeperTestSuite) runCreateGroupTests(poolInfo apptesting.SupportedPoolAndGaugeInfo, concentratedGaugeRecord types.InternalGaugeRecord, balancerGaugeRecord types.InternalGaugeRecord, stableSwapGaugeRecord types.InternalGaugeRecord, expectedCommunityPoolFeeBalance sdk.Coins, expectedGroupGaugeId uint64, originalCommunityPoolBalance sdk.Coins) {
	tests := makeDefaultSuccessCreateGroupTestCases(poolInfo, concentratedGaugeRecord, balancerGaugeRecord, stableSwapGaugeRecord)
	tests = append(tests, []createGroupTestCase{

		{
			name:             "error: no volume in one of the pools",
			coins:            defaultCoins,
			numEpochPaidOver: types.PerpetualNumEpochsPaidOver,
			poolIDs:          []uint64{poolInfo.BalancerPoolID, poolInfo.ConcentratedPoolID},

			// Note that second pool has zero volume
			poolVolumesToSet: []osmomath.Int{defaultVolumeAmount, osmomath.ZeroInt()},

			expectErr: types.NoPoolVolumeError{PoolId: poolInfo.ConcentratedPoolID},
		},
	}...)
	// error cases
	tests = append(tests, makeDefaultErrorCases(poolInfo)...)

	for _, tc := range tests {
		s.Run(tc.name, func() {

			// Ensure we configured volumes and pools correctly
			s.overwriteVolumes(tc.poolIDs, tc.poolVolumesToSet)

			// Since we expect weight syncing to occur, we update the expected weights
			// with the volumes we set above.
			configureExpectedWeights(&tc)

			// Always fund the account with fullyFundedAddressIndex
			s.FundAcc(s.TestAccs[fullyFundedAddressIndex], tc.coins.Add(customGroupCreationFee...))

			groupGaugeId, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, tc.coins, tc.numEpochPaidOver, s.TestAccs[tc.creatorAddressIndex], tc.poolIDs)

			if tc.expectErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectErr.Error())
			} else {
				// For every no error case, increase expected community pool fee balance
				expectedCommunityPoolFeeBalance = expectedCommunityPoolFeeBalance.Add(customGroupCreationFee...)

				s.Require().NoError(err)
				s.Require().Equal(expectedGroupGaugeId, groupGaugeId)

				// Validate group's Gauge
				s.validateGauge(types.Gauge{
					Id:                groupGaugeId,
					Coins:             tc.coins,
					NumEpochsPaidOver: tc.numEpochPaidOver,
					IsPerpetual:       tc.expectedPerpeutalGroupGauge,
					DistributeTo:      incentiveskeeper.ByGroupQueryCondition,
					StartTime:         defaultTime,
				})

				// Validate Group
				s.validateGroupInState(types.Group{
					GroupGaugeId:      expectedGroupGaugeId,
					SplittingPolicy:   types.ByVolume,
					InternalGaugeInfo: tc.expectedGaugeInfo,
				})

				// Validate that community pool was funded with the group creation fee
				s.validateCommunityPoolBalanceUpdatedBy(expectedCommunityPoolFeeBalance, originalCommunityPoolBalance)

				// Bump up expected gauge ID since we are reusing the same test state
				expectedGroupGaugeId++
			}
		})
	}
}

func (s *KeeperTestSuite) runCreateGroupInternalTests(poolInfo apptesting.SupportedPoolAndGaugeInfo, concentratedGaugeRecord types.InternalGaugeRecord, balancerGaugeRecord types.InternalGaugeRecord, stableSwapGaugeRecord types.InternalGaugeRecord, expectedCommunityPoolFeeBalance sdk.Coins, expectedGroupGaugeId uint64, originalCommunityPoolBalance sdk.Coins) {
	tests := makeDefaultSuccessCreateGroupTestCases(poolInfo, concentratedGaugeRecord, balancerGaugeRecord, stableSwapGaugeRecord)
	tests = append(tests, []createGroupTestCase{
		{
			name:                "no volume in one of the pools - does not error due to not attempting to sync",
			coins:               defaultCoins,
			numEpochPaidOver:    types.PerpetualNumEpochsPaidOver,
			creatorAddressIndex: fullyFundedAddressIndex,
			poolIDs:             []uint64{poolInfo.BalancerPoolID, poolInfo.ConcentratedPoolID},
			// Note that second pool has zero volume
			poolVolumesToSet:            []osmomath.Int{defaultVolumeAmount, osmomath.ZeroInt()},
			expectedPerpeutalGroupGauge: true,
			expectedGaugeInfo: addGaugeRecords(defaultEmptyGaugeInfo, []types.InternalGaugeRecord{
				balancerGaugeRecord,
				concentratedGaugeRecord,
			}),
		},
	}...)
	tests = append(tests, makeDefaultErrorCases(poolInfo)...)

	for _, tc := range tests {
		s.Run(tc.name, func() {

			// Ensure we configured volumes and pools correctly
			s.overwriteVolumes(tc.poolIDs, tc.poolVolumesToSet)

			// Always fund the account with fullyFundedAddressIndex
			s.FundAcc(s.TestAccs[fullyFundedAddressIndex], tc.coins.Add(customGroupCreationFee...))

			groupReturn, err := s.App.IncentivesKeeper.CreateGroupInternal(s.Ctx, tc.coins, tc.numEpochPaidOver, s.TestAccs[tc.creatorAddressIndex], tc.poolIDs)
			if tc.expectErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectErr.Error())
			} else {
				// For every no error case, increase expected community pool fee balance
				expectedCommunityPoolFeeBalance = expectedCommunityPoolFeeBalance.Add(customGroupCreationFee...)

				s.Require().NoError(err)

				s.Require().Equal(expectedGroupGaugeId, groupReturn.GroupGaugeId)

				// Validate group's Gauge
				s.validateGauge(types.Gauge{
					Id:                groupReturn.GroupGaugeId,
					Coins:             tc.coins,
					NumEpochsPaidOver: tc.numEpochPaidOver,
					IsPerpetual:       tc.expectedPerpeutalGroupGauge,
					DistributeTo:      incentiveskeeper.ByGroupQueryCondition,
					StartTime:         defaultTime,
				})

				// Attempt to get Group from state and validate that does not exist
				_, err = s.App.IncentivesKeeper.GetGroupByGaugeID(s.Ctx, groupReturn.GroupGaugeId)
				s.Require().Error(err)
				s.Require().ErrorIs(err, types.GroupNotFoundError{GroupGaugeId: groupReturn.GroupGaugeId})

				// Validate Group return
				expectedGroup := types.Group{
					GroupGaugeId:      expectedGroupGaugeId,
					SplittingPolicy:   types.ByVolume,
					InternalGaugeInfo: tc.expectedGaugeInfo,
				}
				s.validateGroup(expectedGroup, groupReturn)

				// Validate that community pool was funded with the group creation fee
				s.validateCommunityPoolBalanceUpdatedBy(expectedCommunityPoolFeeBalance, originalCommunityPoolBalance)

				// Bump up expected gauge ID since we are reusing the same test state
				expectedGroupGaugeId++
			}
		})
	}
}

func (s *KeeperTestSuite) runCreateGroupAsIncentivesModuleAccTests(poolInfo apptesting.SupportedPoolAndGaugeInfo, concentratedGaugeRecord types.InternalGaugeRecord, balancerGaugeRecord types.InternalGaugeRecord, expectedCommunityPoolFeeBalance sdk.Coins, expectedGroupGaugeId uint64, originalCommunityPoolBalance sdk.Coins) {
	tests := []createGroupTestCase{
		{
			name:             "two pools - created perpetual group gauge",
			numEpochPaidOver: types.PerpetualNumEpochsPaidOver,
			poolIDs:          []uint64{poolInfo.ConcentratedPoolID, poolInfo.BalancerPoolID},
			poolVolumesToSet: []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount.Add(defaultVolumeAmount)},

			expectedPerpeutalGroupGauge: true,
			expectedGaugeInfo: addGaugeRecords(defaultEmptyGaugeInfo, []types.InternalGaugeRecord{
				concentratedGaugeRecord,
				balancerGaugeRecord,
			}),
		},
		{
			name:             "no volume in one of the pools - does not error due to not attempting to sync",
			numEpochPaidOver: types.PerpetualNumEpochsPaidOver,
			poolIDs:          []uint64{poolInfo.BalancerPoolID, poolInfo.ConcentratedPoolID},
			// Note that second pool has zero volume
			poolVolumesToSet:            []osmomath.Int{defaultVolumeAmount, osmomath.ZeroInt()},
			expectedPerpeutalGroupGauge: true,
			expectedGaugeInfo: addGaugeRecords(defaultEmptyGaugeInfo, []types.InternalGaugeRecord{
				balancerGaugeRecord,
				concentratedGaugeRecord,
			}),
		},
	}
	// Note: only testing cosmwasm pool error test case (index zero)
	// The rest do not apply because incentive module account is the creator.
	tests = append(tests, makeDefaultErrorCases(poolInfo)[0])

	for _, tc := range tests {
		s.Run(tc.name, func() {

			// Ensure we configured volumes and pools correctly
			s.overwriteVolumes(tc.poolIDs, tc.poolVolumesToSet)

			groupGaugeID, err := s.App.IncentivesKeeper.CreateGroupAsIncentivesModuleAcc(s.Ctx, tc.numEpochPaidOver, tc.poolIDs)
			if tc.expectErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectErr.Error())
			} else {
				// For every no error case, increase expected community pool fee balance
				expectedCommunityPoolFeeBalance = expectedCommunityPoolFeeBalance.Add(customGroupCreationFee...)

				s.Require().NoError(err)

				s.Require().Equal(expectedGroupGaugeId, groupGaugeID)

				// Validate group's Gauge
				s.validateGauge(types.Gauge{
					Id:                groupGaugeID,
					Coins:             tc.coins,
					NumEpochsPaidOver: tc.numEpochPaidOver,
					IsPerpetual:       tc.expectedPerpeutalGroupGauge,
					DistributeTo:      incentiveskeeper.ByGroupQueryCondition,
					StartTime:         defaultTime,
				})

				// Validate Group
				s.validateGroupInState(types.Group{
					GroupGaugeId:      expectedGroupGaugeId,
					SplittingPolicy:   types.ByVolume,
					InternalGaugeInfo: tc.expectedGaugeInfo,
				})

				// Validate that community pool was NOT funded with the group creation fee
				s.validateCommunityPoolBalanceUpdatedBy(emptyCoins, originalCommunityPoolBalance)

				// Bump up expected gauge ID since we are reusing the same test state
				expectedGroupGaugeId++
			}
		})
	}
}

// Validates that the initial gauge info is initialized with the appropriate gauge IDs given pool IDs.
// All weights are set to zero in all cases.
func (s *KeeperTestSuite) TestInitGaugeInfo() {

	// We setup state once for all tests since there are no state mutations
	// in system under test.
	s.SetupTest()
	k := s.App.IncentivesKeeper

	// Prepare pools, their IDs and associated gauge IDs.
	poolInfo := s.PrepareAllSupportedPools()

	// Initialize expected gauge records
	var (
		concentratedGaugeRecord = withRecordGaugeId(defaultZeroWeightGaugeRecord, poolInfo.ConcentratedGaugeID)
		balancerGaugeRecord     = withRecordGaugeId(defaultZeroWeightGaugeRecord, poolInfo.BalancerGaugeID)
		stableSwapGaugeRecord   = withRecordGaugeId(defaultZeroWeightGaugeRecord, poolInfo.StableSwapGaugeID)
	)

	tests := map[string]struct {
		poolIds           []uint64
		expectedGaugeInfo types.InternalGaugeInfo
		expectError       error
	}{
		"one gauge record": {
			poolIds:           []uint64{poolInfo.ConcentratedPoolID},
			expectedGaugeInfo: addGaugeRecords(defaultEmptyGaugeInfo, []types.InternalGaugeRecord{concentratedGaugeRecord}),
		},

		"multiple gauge records": {
			poolIds: []uint64{poolInfo.ConcentratedPoolID, poolInfo.BalancerPoolID, poolInfo.StableSwapPoolID},
			expectedGaugeInfo: addGaugeRecords(defaultEmptyGaugeInfo,
				[]types.InternalGaugeRecord{
					concentratedGaugeRecord,
					balancerGaugeRecord,
					stableSwapGaugeRecord,
				}),
		},

		// error cases

		"error when getting gauge for pool ID (cw pool does not support incentives)": {
			poolIds: []uint64{poolInfo.ConcentratedPoolID, poolInfo.BalancerPoolID, poolInfo.CosmWasmPoolID, poolInfo.StableSwapPoolID},

			expectError: poolincentivetypes.UnsupportedPoolTypeError{PoolID: poolInfo.CosmWasmPoolID, PoolType: poolmanagertypes.CosmWasm},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {

			actualGaugeInfo, err := k.InitGaugeInfo(s.Ctx, tc.poolIds)

			if tc.expectError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// Validate InternalGaugeInfo
			s.validateGaugeInfo(tc.expectedGaugeInfo, actualGaugeInfo)
		})
	}
}

// Tests that group creation fee is charged correctly or bypassed when applicable.
// It can be bypassed if the sender is whitelisted or the sender is the incentives module account.
// Otherwise, the sender must have enough funds to pay the fee.
func (s *KeeperTestSuite) TestChargeGroupCreationFeeIfNotWhitelisted() {
	// Setup test state once at the top and reuse across tests.
	s.SetupTest()

	// Configure group creation fee
	s.App.IncentivesKeeper.SetParam(s.Ctx, types.KeyGroupCreationFee, customGroupCreationFee)

	// Define accounts
	var (
		regularFundedAccount    = s.TestAccs[0]
		regularNotFundedAccount = s.TestAccs[1]
		whitelistedAccount      = s.TestAccs[2]
		incentivesModuleAccount = s.App.AccountKeeper.GetModuleAddress(types.ModuleName)
	)
	// Only fund the regular funded account
	s.FundAcc(regularFundedAccount, customGroupCreationFee)
	defaultWhitelist := []string{whitelistedAccount.String()}

	tests := map[string]struct {
		sender           sdk.AccAddress
		expectFeeCharged bool
		whitelist        []string
		expectError      error
	}{
		"regular funded account - charged": {
			sender:           regularFundedAccount,
			expectFeeCharged: true,
			whitelist:        defaultWhitelist,
		},
		"regular non funded account - error returned": {
			sender:      regularNotFundedAccount,
			whitelist:   defaultWhitelist,
			expectError: errorNoCustomFeeInBalance,
		},
		"unrestricted whitelisted - not charged": {
			sender:    whitelistedAccount,
			whitelist: defaultWhitelist,
		},
		"incentive module account - not charged": {
			sender:    incentivesModuleAccount,
			whitelist: defaultWhitelist,
		},
		"incentive module account with no whitelist - not charged": {
			sender:    incentivesModuleAccount,
			whitelist: []string{},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			incentivesKeeper := s.App.IncentivesKeeper

			// Set up whitelist
			s.App.IncentivesKeeper.SetParam(s.Ctx, types.KeyCreatorWhitelist, tc.whitelist)

			// Keep original balances for final balance assertions
			senderBalanceBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, tc.sender)
			communityPoolBalanceBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName))

			didChargeFee, err := incentivesKeeper.ChargeGroupCreationFeeIfNotWhitelisted(s.Ctx, tc.sender)

			if tc.expectError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expectFeeCharged, didChargeFee)

			senderBalanceAfter := s.App.BankKeeper.GetAllBalances(s.Ctx, tc.sender)
			communityPoolBalanceAfter := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName))

			// Validate balance updates.
			if didChargeFee {
				s.Require().Equal(senderBalanceBefore.Sub(customGroupCreationFee...).String(), senderBalanceAfter.String())
				s.Require().Equal(communityPoolBalanceBefore.Add(customGroupCreationFee...).String(), communityPoolBalanceAfter.String())
			} else {
				s.Require().Equal(senderBalanceBefore.String(), senderBalanceAfter.String())
				s.Require().Equal(communityPoolBalanceBefore.String(), communityPoolBalanceAfter.String())
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetPoolIdsAndDurationsFromGaugeRecords() {
	s.SetupTest()

	poolGroupA := s.PrepareAllSupportedPools()
	poolGroupB := s.PrepareAllSupportedPools()
	poolGroupC := s.PrepareAllSupportedPools()

	longestLockDuration, err := s.App.PoolIncentivesKeeper.GetLongestLockableDuration(s.Ctx)
	s.Require().NoError(err)

	epochDuration := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration

	testCases := []struct {
		name              string
		gaugeRecords      []types.InternalGaugeRecord
		expectedPoolIds   []uint64
		expectedDurations []time.Duration
		expectError       error
	}{
		{
			name: "group consisting of Balancer, Concentrated, and StableSwap pools",
			gaugeRecords: []types.InternalGaugeRecord{
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupA.BalancerGaugeID),
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupA.ConcentratedGaugeID),
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupA.StableSwapGaugeID)},
			expectedPoolIds:   []uint64{poolGroupA.BalancerPoolID, poolGroupA.ConcentratedPoolID, poolGroupA.StableSwapPoolID},
			expectedDurations: []time.Duration{longestLockDuration, epochDuration, longestLockDuration},
		},
		{
			name: "group consisting of only Balancer pools",
			gaugeRecords: []types.InternalGaugeRecord{
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupA.BalancerGaugeID),
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupB.BalancerGaugeID),
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupC.BalancerGaugeID)},
			expectedPoolIds:   []uint64{poolGroupA.BalancerPoolID, poolGroupB.BalancerPoolID, poolGroupC.BalancerPoolID},
			expectedDurations: []time.Duration{longestLockDuration, longestLockDuration, longestLockDuration},
		},
		{
			name: "group consisting of only Concentrated pools",
			gaugeRecords: []types.InternalGaugeRecord{
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupA.ConcentratedGaugeID),
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupB.ConcentratedGaugeID),
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupC.ConcentratedGaugeID)},
			expectedPoolIds:   []uint64{poolGroupA.ConcentratedPoolID, poolGroupB.ConcentratedPoolID, poolGroupC.ConcentratedPoolID},
			expectedDurations: []time.Duration{epochDuration, epochDuration, epochDuration},
		},
		{
			name: "error: one of the gauge records has a non-existent gauge ID",
			gaugeRecords: []types.InternalGaugeRecord{
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupA.BalancerGaugeID),
				withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroupA.ConcentratedGaugeID),
				withRecordGaugeId(defaultZeroWeightGaugeRecord, 0)},
			expectError: types.GaugeNotFoundError{GaugeID: 0},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// system under test
			poolIds, durations, err := s.App.IncentivesKeeper.GetPoolIdsAndDurationsFromGaugeRecords(s.Ctx, tc.gaugeRecords)
			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedPoolIds, poolIds)
			s.Require().Equal(tc.expectedDurations, durations)
		})
	}
}

func (s *KeeperTestSuite) TestGetPoolIdAndDurationFromGaugeRecord() {
	s.SetupTest()

	poolGroup := s.PrepareAllSupportedPools()

	longestLockDuration, err := s.App.PoolIncentivesKeeper.GetLongestLockableDuration(s.Ctx)
	s.Require().NoError(err)

	epochDuration := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration

	testCases := []struct {
		name             string
		gaugeRecord      types.InternalGaugeRecord
		expectedPoolId   uint64
		expectedDuration time.Duration
		expectedErr      error
	}{
		{
			name:             "balancer pool record",
			gaugeRecord:      withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroup.BalancerGaugeID),
			expectedPoolId:   poolGroup.BalancerPoolID,
			expectedDuration: longestLockDuration,
		},
		{
			name:             "concentrated pool record",
			gaugeRecord:      withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroup.ConcentratedGaugeID),
			expectedPoolId:   poolGroup.ConcentratedPoolID,
			expectedDuration: epochDuration,
		},
		{
			name:             "stable swap pool record",
			gaugeRecord:      withRecordGaugeId(defaultZeroWeightGaugeRecord, poolGroup.StableSwapGaugeID),
			expectedPoolId:   poolGroup.StableSwapPoolID,
			expectedDuration: longestLockDuration,
		},
		{
			name:        "err: non-existent gauge ID",
			gaugeRecord: withRecordGaugeId(defaultZeroWeightGaugeRecord, 0),
			expectedErr: types.GaugeNotFoundError{GaugeID: 0},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// system under test
			poolId, duration, err := s.App.IncentivesKeeper.GetPoolIdAndDurationFromGaugeRecord(s.Ctx, tc.gaugeRecord)
			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectedErr)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedPoolId, poolId)
			s.Require().Equal(tc.expectedDuration, duration)
		})
	}
}

// validates that the actual group in state equals the expected group
func (s *KeeperTestSuite) validateGroupInState(expectedGroup types.Group) {
	actualGroup, err := s.App.IncentivesKeeper.GetGroupByGaugeID(s.Ctx, expectedGroup.GroupGaugeId)
	s.Require().NoError(err)
	s.validateGroup(expectedGroup, actualGroup)
}

// validates that community pool was funded with the given amount
func (s *KeeperTestSuite) validateCommunityPoolBalanceUpdatedBy(expectedCoinUpdate, originalCommunityPoolBalance sdk.Coins) {
	communityPoolAddress := s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	communityPoolBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)
	s.Require().Equal(expectedCoinUpdate.String(), communityPoolBalance.Sub(originalCommunityPoolBalance...).String())
}

// validates that the given actual group equals the expected group
func (s *KeeperTestSuite) validateGroup(expectedGroup types.Group, actualGroup types.Group) {
	s.Require().Equal(expectedGroup.GroupGaugeId, actualGroup.GroupGaugeId)
	s.Require().Equal(expectedGroup.SplittingPolicy, actualGroup.SplittingPolicy)
	s.validateGaugeInfo(expectedGroup.InternalGaugeInfo, actualGroup.InternalGaugeInfo)
}

// validates that Group and group Gauge do not exist
func (s *KeeperTestSuite) validateGroupNotExists(nonPerpetualGroupGaugeID uint64) {
	_, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, nonPerpetualGroupGaugeID)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.GaugeNotFoundError{GaugeID: nonPerpetualGroupGaugeID})

	_, err = s.App.IncentivesKeeper.GetGroupByGaugeID(s.Ctx, nonPerpetualGroupGaugeID)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.GroupNotFoundError{GroupGaugeId: nonPerpetualGroupGaugeID})
}

// configures expected weights on the test case
func configureExpectedWeights(tc *createGroupTestCase) {
	expectedTotalVolume := osmomath.ZeroInt()
	for i := range tc.poolIDs {

		if tc.expectedGaugeInfo.GaugeRecords == nil {
			continue
		}

		tc.expectedGaugeInfo.GaugeRecords[i].CumulativeWeight = tc.poolVolumesToSet[i]
		tc.expectedGaugeInfo.GaugeRecords[i].CurrentWeight = tc.poolVolumesToSet[i]

		expectedTotalVolume = expectedTotalVolume.Add(tc.poolVolumesToSet[i])
	}
	tc.expectedGaugeInfo.TotalWeight = expectedTotalVolume
}
