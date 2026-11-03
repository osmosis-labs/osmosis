package v33_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	minttypes "github.com/osmosis-labs/osmosis/v32/x/mint/types"
)

var (
	// Mainnet mint state before the v33 upgrade.
	mainnetEpochProvisions         = osmomath.MustNewDecFromStr("121765601217.656011811263318112")
	mainnetDistributionProportions = minttypes.DistributionProportions{
		Staking:          osmomath.MustNewDecFromStr("0.08"),
		PoolIncentives:   osmomath.ZeroDec(),
		DeveloperRewards: osmomath.MustNewDecFromStr("0.25"),
		CommunityPool:    osmomath.MustNewDecFromStr("0.67"),
	}
	devRewardsReceiver = sdk.AccAddress([]byte("dev_rewards_receiver"))
)

// TestRemoveDeveloperRewardsFromMint runs the v33 upgrade on the mainnet mint state and checks that the
// developer rewards share is removed without changing what each recipient receives per epoch.
func (s *UpgradeTestSuite) TestRemoveDeveloperRewardsFromMint() {
	s.prepareMainnetMintState()
	mintDenom := s.App.MintKeeper.GetParams(s.Ctx).MintDenom

	// Per-epoch amounts under the old flow: the full provisions are minted, the developer share is
	// burned and paid out of the developer vesting module account instead.
	oldMinted := mainnetEpochProvisions.TruncateInt()
	oldStaking := oldMinted.ToLegacyDec().Mul(mainnetDistributionProportions.Staking).TruncateInt()
	oldDevRewards := oldMinted.ToLegacyDec().Mul(mainnetDistributionProportions.DeveloperRewards).TruncateInt()
	oldCommunityPool := oldMinted.Sub(oldStaking).Sub(oldDevRewards)
	oldNewSupply := oldMinted.Sub(oldDevRewards)

	s.runUpgrade()

	params := s.App.MintKeeper.GetParams(s.Ctx)
	proportions := params.DistributionProportions
	s.Require().True(proportions.DeveloperRewards.IsZero())
	s.Require().Empty(params.WeightedDeveloperRewardsReceivers)
	s.Require().Equal(osmomath.MustNewDecFromStr("0.106666666666666667"), proportions.Staking)
	s.Require().True(proportions.PoolIncentives.IsZero())
	s.Require().Equal(osmomath.MustNewDecFromStr("0.893333333333333333"), proportions.CommunityPool)
	s.Require().Equal(osmomath.OneDec(), proportions.Staking.Add(proportions.PoolIncentives).Add(proportions.CommunityPool))
	s.Require().NoError(params.Validate())

	s.Require().Equal(mainnetEpochProvisions.Mul(osmomath.MustNewDecFromStr("0.75")), s.App.MintKeeper.GetMinter(s.Ctx).EpochProvisions)

	// Run one mint epoch on the new state.
	feeCollector := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	communityPool := s.App.AccountKeeper.GetModuleAddress(distributiontypes.ModuleName)
	devVesting := s.App.AccountKeeper.GetModuleAddress(minttypes.DeveloperVestingModuleAcctName)

	stakingBefore := s.App.BankKeeper.GetBalance(s.Ctx, feeCollector, mintDenom).Amount
	communityPoolBefore := s.App.BankKeeper.GetBalance(s.Ctx, communityPool, mintDenom).Amount
	devVestingBefore := s.App.BankKeeper.GetBalance(s.Ctx, devVesting, mintDenom)
	supplyBefore := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, mintDenom).Amount
	supplyOffsetBefore := s.App.BankKeeper.GetSupplyOffset(s.Ctx, mintDenom)

	s.runMintEpoch()

	staking := s.App.BankKeeper.GetBalance(s.Ctx, feeCollector, mintDenom).Amount.Sub(stakingBefore)
	community := s.App.BankKeeper.GetBalance(s.Ctx, communityPool, mintDenom).Amount.Sub(communityPoolBefore)
	newSupply := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, mintDenom).Amount.Sub(supplyBefore)

	// Rescaling the proportions rounds at 18 decimals, so allow a 1 uosmo difference per recipient.
	oneUnit := osmomath.OneInt()
	s.Require().True(staking.Sub(oldStaking).Abs().LTE(oneUnit), "staking: old %s, new %s", oldStaking, staking)
	s.Require().True(community.Sub(oldCommunityPool).Abs().LTE(oneUnit), "community pool: old %s, new %s", oldCommunityPool, community)
	s.Require().True(newSupply.Sub(oldNewSupply).Abs().LTE(oneUnit), "new supply: old %s, new %s", oldNewSupply, newSupply)

	// The developer vesting module account, the supply offset and the old receivers are untouched.
	s.Require().Equal(devVestingBefore, s.App.BankKeeper.GetBalance(s.Ctx, devVesting, mintDenom))
	s.Require().Equal(supplyOffsetBefore, s.App.BankKeeper.GetSupplyOffset(s.Ctx, mintDenom))
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, devRewardsReceiver, mintDenom).IsZero())
}

// TestRemoveDeveloperRewardsFromMintNoop checks that the upgrade leaves the mint state alone when the
// developer rewards share is already zero.
func (s *UpgradeTestSuite) TestRemoveDeveloperRewardsFromMintNoop() {
	paramsBefore := s.App.MintKeeper.GetParams(s.Ctx)
	s.Require().True(paramsBefore.DistributionProportions.DeveloperRewards.IsZero())
	minterBefore := s.App.MintKeeper.GetMinter(s.Ctx)

	s.runUpgrade()

	s.Require().Equal(paramsBefore, s.App.MintKeeper.GetParams(s.Ctx))
	s.Require().Equal(minterBefore, s.App.MintKeeper.GetMinter(s.Ctx))
}

// prepareMainnetMintState writes the pre-upgrade mainnet mint params and minter. The params are written
// key by key because the non-zero developer rewards proportion no longer passes params validation.
func (s *UpgradeTestSuite) prepareMainnetMintState() {
	s.App.MintKeeper.SetParam(s.Ctx, minttypes.KeyPoolAllocationRatio, mainnetDistributionProportions)
	s.App.MintKeeper.SetParam(s.Ctx, minttypes.KeyDeveloperRewardsReceiver, []minttypes.WeightedAddress{
		{Address: devRewardsReceiver.String(), Weight: osmomath.OneDec()},
	})
	s.App.MintKeeper.SetMinter(s.Ctx, minttypes.NewMinter(mainnetEpochProvisions))
}

// runMintEpoch ends the distribution start epoch. The start epoch resets the last reduction epoch,
// so no reduction is triggered.
func (s *UpgradeTestSuite) runMintEpoch() {
	params := s.App.MintKeeper.GetParams(s.Ctx)
	s.Require().Positive(params.ReductionPeriodInEpochs)
	s.Require().NoError(s.App.MintKeeper.AfterEpochEnd(s.Ctx, params.EpochIdentifier, params.MintingRewardsDistributionStartEpoch))
}
