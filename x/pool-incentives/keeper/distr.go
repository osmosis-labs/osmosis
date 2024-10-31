package keeper

import (
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FundCommunityPoolFromModule allows the pool-incentives module to directly fund the community fund pool.
func (k Keeper) FundCommunityPoolFromModule(ctx sdk.Context, asset sdk.Coin) error {
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		panic("Could not get distribution module from SDK")
	}

	return k.distrKeeper.FundCommunityPool(ctx, sdk.Coins{asset}, moduleAddr)
}

// AllocateAsset allocates and distributes coin according a gauge’s proportional weight that is recorded in the record.
func (k Keeper) AllocateAsset(ctx sdk.Context) error {
	logger := k.Logger(ctx)
	params := k.GetParams(ctx)
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	asset := k.bankKeeper.GetBalance(ctx, moduleAddr, params.MintedDenom)
	if asset.Amount.IsZero() {
		// when allocating asset is zero, skip execution
		return nil
	}

	distrInfo := k.GetDistrInfo(ctx)

	if distrInfo.TotalWeight.IsZero() {
		// If there are no records, put the asset to the community pool
		return k.FundCommunityPoolFromModule(ctx, asset)
	}

	ctx.Logger().Info("AllocateAsset minted amount", "module", types.ModuleName, "totalMintedAmount", asset.Amount, "height", ctx.BlockHeight())

	assetAmountDec := asset.Amount.ToLegacyDec()
	totalWeightDec := distrInfo.TotalWeight.ToLegacyDec()
	for _, record := range distrInfo.Records {
		allocatingAmount := assetAmountDec.Mul(record.Weight.ToLegacyDec().Quo(totalWeightDec)).TruncateInt()

		// when weight is too small and no amount is allocated, just skip this to avoid zero coin send issues
		if !allocatingAmount.IsPositive() {
			logger.Info(fmt.Sprintf("allocating amount for (%d, %s) record is not positive", record.GaugeId, record.Weight.String()))
			continue
		}

		if record.GaugeId == types.CommunityPoolDistributionGaugeID { // fund community pool if gaugeId is zero
			if err := k.FundCommunityPoolFromModule(ctx, sdk.NewCoin(asset.Denom, allocatingAmount)); err != nil {
				return err
			}
			continue
		}

		coins := sdk.NewCoins(sdk.NewCoin(asset.Denom, allocatingAmount))
		ctx.Logger().Debug("Adding to gauge rewards", "module", types.ModuleName, "gaugeId", record.GaugeId, "coins", coins.String(), "height", ctx.BlockHeight())
		err := k.incentivesKeeper.AddToGaugeRewards(ctx, k.accountKeeper.GetModuleAddress(types.ModuleName), coins, record.GaugeId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) GetDistrInfo(ctx sdk.Context) types.DistrInfo {
	store := ctx.KVStore(k.storeKey)
	distrInfo := types.DistrInfo{}
	osmoutils.MustGet(store, types.DistrInfoKey, &distrInfo)
	return distrInfo
}

func (k Keeper) SetDistrInfo(ctx sdk.Context, distrInfo types.DistrInfo) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.DistrInfoKey, &distrInfo)
}

// validateRecords validates a list of records to ensure that:
// 1) there are no duplicates,
// 2) the records are in sorted order.
// 3) the records only pay to gauges that exist.
func (k Keeper) validateRecords(ctx sdk.Context, records ...types.DistrRecord) error {
	lastGaugeID := uint64(0)
	gaugeIdFlags := make(map[uint64]bool)

	for _, record := range records {
		if gaugeIdFlags[record.GaugeId] {
			return errorsmod.Wrapf(
				types.ErrDistrRecordRegisteredGauge,
				"Gauge ID #%d has duplications.",
				record.GaugeId,
			)
		}

		// Ensure records are sorted because ~AESTHETIC~
		if record.GaugeId < lastGaugeID {
			return errorsmod.Wrapf(
				types.ErrDistrRecordNotSorted,
				"Gauge ID #%d came after Gauge ID #%d.",
				record.GaugeId, lastGaugeID,
			)
		}
		lastGaugeID = record.GaugeId

		// unless GaugeID is 0 for the community pool, don't allow distribution records for gauges that don't exist
		if record.GaugeId != 0 {
			gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, record.GaugeId)
			if err != nil {
				return err
			}
			if !gauge.IsPerpetual {
				return errorsmod.Wrapf(types.ErrDistrRecordRegisteredGauge,
					"Gauge ID #%d is not perpetual.",
					record.GaugeId)
			}
		}

		gaugeIdFlags[record.GaugeId] = true
	}
	return nil
}

// This is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) ReplaceDistrRecords(ctx sdk.Context, records ...types.DistrRecord) error {
	distrInfo := k.GetDistrInfo(ctx)

	err := k.validateRecords(ctx, records...)
	if err != nil {
		return err
	}

	totalWeight := osmomath.NewInt(0)

	for _, record := range records {
		totalWeight = totalWeight.Add(record.Weight)
	}

	distrInfo.Records = records
	distrInfo.TotalWeight = totalWeight

	k.SetDistrInfo(ctx, distrInfo)
	return nil
}

// UpdateDistrRecords is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) UpdateDistrRecords(ctx sdk.Context, records ...types.DistrRecord) error {
	recordsMap := make(map[uint64]types.DistrRecord)
	totalWeight := osmomath.NewInt(0)

	for _, existingRecord := range k.GetDistrInfo(ctx).Records {
		recordsMap[existingRecord.GaugeId] = existingRecord
		totalWeight = totalWeight.Add(existingRecord.Weight)
	}

	err := k.validateRecords(ctx, records...)
	if err != nil {
		return err
	}

	for _, record := range records {
		if val, ok := recordsMap[record.GaugeId]; ok {
			totalWeight = totalWeight.Sub(val.Weight)
			recordsMap[record.GaugeId] = record
			totalWeight = totalWeight.Add(record.Weight)
		} else {
			recordsMap[record.GaugeId] = record
			totalWeight = totalWeight.Add(record.Weight)
		}
	}

	newRecords := []types.DistrRecord{}

	for _, val := range recordsMap {
		if !val.Weight.Equal(osmomath.ZeroInt()) {
			newRecords = append(newRecords, val)
		}
	}

	sort.SliceStable(newRecords, func(i, j int) bool {
		return newRecords[i].GaugeId < newRecords[j].GaugeId
	})

	k.SetDistrInfo(ctx, types.DistrInfo{
		Records:     newRecords,
		TotalWeight: totalWeight,
	})
	return nil
}
