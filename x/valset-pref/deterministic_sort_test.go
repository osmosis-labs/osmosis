package keeper_test

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	valPref "github.com/osmosis-labs/osmosis/v32/x/valset-pref"
)

// tiedValRatios returns n ValRatios where VRatio takes only three values, so most entries tie.
// ValAddr encodes the input index so the test can check where each entry ends up.
func tiedValRatios(n int) []valPref.ValRatio {
	ratios := make([]valPref.ValRatio, n)
	for i := range ratios {
		ratios[i] = valPref.ValRatio{
			ValAddr: sdk.ValAddress([]byte{byte(i)}),
			VRatio:  osmomath.NewDecWithPrec(int64((i*7)%3+1), 1),
		}
	}
	return ratios
}

func (s *KeeperTestSuite) TestSortValRatiosDesc() {
	const n = 64

	s.Run("from the gate height: descending VRatio, ties keep input order", func() {
		ctx := s.Ctx.WithBlockHeight(valPref.DeterministicValRatioSortHeight)
		ratios := tiedValRatios(n)
		valPref.SortValRatiosDesc(ctx, ratios)

		for i := 1; i < n; i++ {
			prev, cur := ratios[i-1], ratios[i]
			s.Require().True(prev.VRatio.GTE(cur.VRatio), "not descending at index %d", i)
			if prev.VRatio.Equal(cur.VRatio) {
				s.Require().Less(prev.ValAddr[0], cur.ValAddr[0], "tie at index %d did not keep input order", i)
			}
		}
	})

	s.Run("below the gate height: same order as sort.Slice", func() {
		ctx := s.Ctx.WithBlockHeight(valPref.DeterministicValRatioSortHeight - 1)
		ratios := tiedValRatios(n)
		expected := tiedValRatios(n)
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].VRatio.GT(expected[j].VRatio)
		})

		valPref.SortValRatiosDesc(ctx, ratios)
		s.Require().Equal(expected, ratios)
	})
}

// TestUndelegateFromRebalancedValidatorSetTieOrder reproduces the shape of the tx that halted
// osmosis-1 at height 74,955,153: a delegator with no valset preference whose delegations give
// many tied VRatios, in an order that is not already sorted. Delegations of 100, 200 and 300
// with a total undelegation of 1/6 of the stake truncate to 16, 33 and 50, so VRatio takes
// three values (0.16, 0.165, 0.1666...) and each value is shared by a third of the validators.
// From the gate height, ties must keep the store order of the delegations, which fixes the order
// of the DVPairs in the unbonding queue.
func (s *KeeperTestSuite) TestUndelegateFromRebalancedValidatorSetTieOrder() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(valPref.DeterministicValRatioSortHeight)

	const numVals = 30
	valAddrs := s.SetupMultipleValidators(numVals)
	delegator := s.TestAccs[0]
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)

	amtByVal := map[string]osmomath.Int{}
	total := osmomath.ZeroInt()
	for i, valAddr := range valAddrs {
		amt := osmomath.NewInt(int64(100 * (i%3 + 1)))
		amtByVal[valAddr] = amt
		total = total.Add(amt)
		s.FundAcc(delegator, sdk.NewCoins(sdk.NewCoin(bondDenom, amt)))
		s.Require().NoError(s.PrepareExistingDelegations(s.Ctx, []string{valAddr}, delegator, amt))
	}

	// store order of the delegations, then a stable sort by delegation size descending
	// (larger delegation means larger VRatio here)
	delegations, err := s.App.StakingKeeper.GetDelegatorDelegations(s.Ctx, delegator, numVals)
	s.Require().NoError(err)
	s.Require().Len(delegations, numVals)
	expectedOrder := make([]string, 0, numVals)
	for _, d := range delegations {
		expectedOrder = append(expectedOrder, d.ValidatorAddress)
	}
	sort.SliceStable(expectedOrder, func(i, j int) bool {
		return amtByVal[expectedOrder[i]].GT(amtByVal[expectedOrder[j]])
	})

	err = s.App.ValidatorSetPreferenceKeeper.UndelegateFromRebalancedValidatorSet(
		s.Ctx, delegator.String(), sdk.NewCoin(bondDenom, total.QuoRaw(6)))
	s.Require().NoError(err)

	unbondingTime, err := s.App.StakingKeeper.UnbondingTime(s.Ctx)
	s.Require().NoError(err)
	dvPairs, err := s.App.StakingKeeper.GetUBDQueueTimeSlice(s.Ctx, s.Ctx.BlockHeader().Time.Add(unbondingTime))
	s.Require().NoError(err)
	s.Require().Len(dvPairs, numVals)

	for i, valAddr := range expectedOrder {
		s.Require().Equal(valAddr, dvPairs[i].ValidatorAddress, "unbonding queue order differs at index %d", i)
		s.Require().Equal(delegator.String(), dvPairs[i].DelegatorAddress)
	}
}

// TestSortValRatiosDescHaltBlock pins the order for the tx in osmosis-1 block 74,955,152,
// using the VRatios computed from chain state at 74,955,151 (sender osmo1v5prcj2fsk23qft97rhhyfjh8wtg35vldrl3cu,
// undelegating 122157066uosmo across 15 delegations, five of them fully, so VRatio = 1 ties).
// Block 74,955,152 is below the gate, so it must reproduce the order sort.Slice gives on Go 1.24+,
// which is what the official v32.0.0 release binaries committed. A Go 1.23 build gives a different
// order for the five ties and fails this test.
func (s *KeeperTestSuite) TestSortValRatiosDescHaltBlock() {
	// store order of the delegations (by validator address bytes)
	inputs := []struct{ name, vRatio string }{
		{"polkachu", "0.999999151919249143"},
		{"Lavender.Five", "1.000000000000000000"},
		{"Bro_n_Bro", "0.999999965048538667"},
		{"IcyCRO", "0.999949634852681944"},
		{"NodeGuardians", "0.999999897435750164"},
		{"Binary Builders", "0.999999904575827626"},
		{"SHUTDOWN | REDELEGATE", "1.000000000000000000"},
		{"Do Not Delegate", "0.999854904236796286"},
		{"Allnodes", "0.999999956458784212"},
		{"SmartNodes", "1.000000000000000000"},
		{"WhisperNode", "1.000000000000000000"},
		{"Oldcat", "0.999999791546110834"},
		{"ECO Stake", "1.000000000000000000"},
		{"Shutting down, please redelegate", "0.999999797347932929"},
		{"Shutdown - please redelegate", "0.999999766512853117"},
	}
	ratios := func() []valPref.ValRatio {
		out := make([]valPref.ValRatio, len(inputs))
		for i, in := range inputs {
			out[i] = valPref.ValRatio{ValAddr: sdk.ValAddress([]byte{byte(i)}), VRatio: osmomath.MustNewDecFromStr(in.vRatio)}
		}
		return out
	}
	names := func(rs []valPref.ValRatio) []string {
		out := make([]string, len(rs))
		for i, r := range rs {
			out[i] = inputs[r.ValAddr[0]].name
		}
		return out
	}
	tail := []string{
		"Bro_n_Bro", "Allnodes", "Binary Builders", "NodeGuardians", "Shutting down, please redelegate",
		"Oldcat", "Shutdown - please redelegate", "polkachu", "IcyCRO", "Do Not Delegate",
	}

	s.Run("block 74,955,152: Go 1.24+ sort.Slice order", func() {
		rs := ratios()
		valPref.SortValRatiosDesc(s.Ctx.WithBlockHeight(74_955_152), rs)
		expected := append([]string{"SmartNodes", "SHUTDOWN | REDELEGATE", "ECO Stake", "Lavender.Five", "WhisperNode"}, tail...)
		s.Require().Equal(expected, names(rs))
	})

	s.Run("same input at the gate height: ties in store order", func() {
		rs := ratios()
		valPref.SortValRatiosDesc(s.Ctx.WithBlockHeight(valPref.DeterministicValRatioSortHeight), rs)
		expected := append([]string{"Lavender.Five", "SHUTDOWN | REDELEGATE", "SmartNodes", "WhisperNode", "ECO Stake"}, tail...)
		s.Require().Equal(expected, names(rs))
	})
}
