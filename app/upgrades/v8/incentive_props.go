package v8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v8/osmoutils"
	poolincentiveskeeper "github.com/osmosis-labs/osmosis/v8/x/pool-incentives/keeper"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v8/x/pool-incentives/types"
)

// This file implements logic for accelerated incentive proposals.

// This function is common to all the props,
// executing the equivalent result of the "UpdatePoolIncentives" proposals, inside of this upgrade logic.
func applyPoolIncentivesUpdate(ctx sdk.Context, poolincentiveskeeper *poolincentiveskeeper.Keeper, records []poolincentivestypes.DistrRecord) {
	// Notice that the pool incentives update proposal code, just calls UpdateDistrRecords:
	// https://github.com/osmosis-labs/osmosis/blob/v7.3.0/x/pool-incentives/keeper/gov.go#L13-L15
	// And that p.Records is the field output by the gov queries.

	// If error, undo state update, log, and proceed. We don't want to stop the entire upgrade due to
	// an unexpected error here.
	_ = osmoutils.ApplyFuncIfNoError(ctx, func(wrappedCtx sdk.Context) error {
		err := poolincentiveskeeper.UpdateDistrRecords(wrappedCtx, records...)
		if err != nil {
			ctx.Logger().Error("Something has happened, prop update did not apply. Continuing to proceed with other components of the upgrade.")
		}
		return err
	})
}

// Apply prop 222 change
func ApplyProp222Change(ctx sdk.Context, poolincentiveskeeper *poolincentiveskeeper.Keeper) {
	// Pool records obtained right off proposal
	// osmosisd q gov proposal 222
	// records:
	// - gauge_id: "1718"
	//   weight: "9138119"
	// - gauge_id: "1719"
	//   weight: "5482871"
	// - gauge_id: "1720"
	//   weight: "3655247"
	// - gauge_id: "2965"
	//   weight: "9138119"
	// - gauge_id: "2966"
	//   weight: "5482872"
	// - gauge_id: "2967"
	//   weight: "3655248"
	// _PLEASE_ double check these numbers, and double the check the proposals choice itself
	records := []poolincentivestypes.DistrRecord{
		{GaugeId: 1718, Weight: sdk.NewInt(9138119)},
		{GaugeId: 1719, Weight: sdk.NewInt(5482871)},
		{GaugeId: 1720, Weight: sdk.NewInt(3655247)},
		{GaugeId: 2965, Weight: sdk.NewInt(9138119)},
		{GaugeId: 2966, Weight: sdk.NewInt(5482872)},
		{GaugeId: 2967, Weight: sdk.NewInt(3655248)},
	}

	ctx.Logger().Info("Applying proposal 222 update")
	applyPoolIncentivesUpdate(ctx, poolincentiveskeeper, records)
}

// Apply prop 223 change
func ApplyProp223Change(ctx sdk.Context, poolincentiveskeeper *poolincentiveskeeper.Keeper) {
	// Pool records obtained right off proposal
	// osmosisd q gov proposal 223
	// records:
	// 	- gauge_id: "1721"
	//     weight: "2831977"
	//   - gauge_id: "1722"
	//     weight: "1699186"
	//   - gauge_id: "1723"
	//     weight: "1132791"
	//   - gauge_id: "3383"
	//     weight: "2831978"
	//   - gauge_id: "3384"
	//     weight: "1699187"
	//   - gauge_id: "3385"
	//     weight: "1132791"

	// _PLEASE_ double check these numbers, and double the check the proposals choice itself
	records := []poolincentivestypes.DistrRecord{
		{GaugeId: 1721, Weight: sdk.NewInt(2831977)},
		{GaugeId: 1722, Weight: sdk.NewInt(1699186)},
		{GaugeId: 1723, Weight: sdk.NewInt(1132791)},
		{GaugeId: 3383, Weight: sdk.NewInt(2831978)},
		{GaugeId: 3384, Weight: sdk.NewInt(1699187)},
		{GaugeId: 3385, Weight: sdk.NewInt(1132791)},
	}

	ctx.Logger().Info("Applying proposal 223 update")
	applyPoolIncentivesUpdate(ctx, poolincentiveskeeper, records)
}

// Apply prop 224 change
func ApplyProp224Change(ctx sdk.Context, poolincentiveskeeper *poolincentiveskeeper.Keeper) {
	// Pool records obtained right off proposal
	// osmosisd q gov proposal 224
	// records:
	// - gauge_id: "1724"
	//   weight: "1881159"
	// - gauge_id: "1725"
	//   weight: "1128695"
	// - gauge_id: "1726"
	//   weight: "752463"
	// - gauge_id: "2949"
	//   weight: "1881160"
	// - gauge_id: "2950"
	//   weight: "1128696"
	// - gauge_id: "2951"
	//   weight: "752464"
	// _PLEASE_ double check these numbers, and double the check the proposals choice itself
	records := []poolincentivestypes.DistrRecord{
		{GaugeId: 1724, Weight: sdk.NewInt(1881159)},
		{GaugeId: 1725, Weight: sdk.NewInt(1128695)},
		{GaugeId: 1726, Weight: sdk.NewInt(752463)},
		{GaugeId: 2949, Weight: sdk.NewInt(1881160)},
		{GaugeId: 2950, Weight: sdk.NewInt(1128696)},
		{GaugeId: 2951, Weight: sdk.NewInt(752464)},
	}

	ctx.Logger().Info("Applying proposal 224 update")
	applyPoolIncentivesUpdate(ctx, poolincentiveskeeper, records)
}
