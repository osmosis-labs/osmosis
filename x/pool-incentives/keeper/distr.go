package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/v3/x/pool-incentives/types"
)

func (k Keeper) FundCommunityPoolFromModule(ctx sdk.Context, asset sdk.Coin) error {
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.communityPoolName, sdk.Coins{asset})
	if err != nil {
		return err
	}

	feePool := k.distrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(asset)...)
	k.distrKeeper.SetFeePool(ctx, feePool)
	return nil
}

// AllocateAsset allocates and distributes coin according a gauge’s proportional weight that is recorded in the record
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

	assetAmountDec := asset.Amount.ToDec()
	totalWeightDec := distrInfo.TotalWeight.ToDec()
	for _, record := range distrInfo.Records {
		allocatingAmount := assetAmountDec.Mul(record.Weight.ToDec().Quo(totalWeightDec)).TruncateInt()

		// when weight is too small and no amount is allocated, just skip this to avoid zero coin send issues
		if !allocatingAmount.IsPositive() {
			logger.Info(fmt.Sprintf("allocating amount for (%d, %s) record is not positive", record.GaugeId, record.Weight.String()))
			continue
		}

		if record.GaugeId == 0 { // fund community pool if gaugeId is zero
			err := k.FundCommunityPoolFromModule(ctx, sdk.NewCoin(asset.Denom, allocatingAmount))
			if err != nil {
				return err
			}
			continue
		}

		coins := sdk.NewCoins(sdk.NewCoin(asset.Denom, allocatingAmount))
		err := k.incentivesKeeper.AddToGaugeRewards(ctx, k.accountKeeper.GetModuleAddress(types.ModuleName), coins, record.GaugeId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) GetDistrInfo(ctx sdk.Context) types.DistrInfo {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DistrInfoKey)

	if len(bz) == 0 {
		distrInfo := types.DistrInfo{
			TotalWeight: sdk.NewInt(0),
			Records:     nil,
		}
		bz = k.cdc.MustMarshalBinaryBare(&distrInfo)

		store.Set(types.DistrInfoKey, bz)
		return distrInfo
	}

	distrInfo := types.DistrInfo{}
	k.cdc.MustUnmarshalBinaryBare(bz, &distrInfo)

	return distrInfo
}

func (k Keeper) SetDistrInfo(ctx sdk.Context, distrInfo types.DistrInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&distrInfo)
	store.Set(types.DistrInfoKey, bz)
}

// indexOfDistrRecordByGaugeId returns the index of the record for the specific gauge id.
// If there is no record matched to the gauge id, return -1.
func (k Keeper) indexOfDistrRecordByGaugeId(ctx sdk.Context, gaugeId uint64) int {
	distrInfo := k.GetDistrInfo(ctx)
	records := distrInfo.Records
	for i, record := range records {
		if record.GaugeId == gaugeId {
			return i
		}
	}
	return -1
}

// This is checked for no err when a proposal is made, and executed when a proposal passes
func (k Keeper) ReplaceDistrRecords(ctx sdk.Context, records ...types.DistrRecord) error {
	distrInfo := k.GetDistrInfo(ctx)

	lastGaugeID := uint64(0)
	gaugeIdFlags := make(map[uint64]bool)

	totalWeight := sdk.NewInt(0)

	for _, record := range records {
		if gaugeIdFlags[record.GaugeId] {
			return sdkerrors.Wrapf(
				types.ErrDistrRecordRegisteredGauge,
				"Gauge ID #%d has duplications.",
				record.GaugeId,
			)
		}

		// Ensure records are sorted because ~AESTHETIC~
		if record.GaugeId < lastGaugeID {
			return sdkerrors.Wrapf(
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
				return sdkerrors.Wrapf(types.ErrDistrRecordRegisteredGauge,
					"Gauge ID #%d is not perpetual.",
					record.GaugeId)
			}
		}

		gaugeIdFlags[record.GaugeId] = true
		totalWeight = totalWeight.Add(record.Weight)
	}

	distrInfo.Records = records
	distrInfo.TotalWeight = totalWeight

	k.SetDistrInfo(ctx, distrInfo)
	return nil
}

// This is checked for no err when a proposal is made, and executed when a proposal passes
func (k Keeper) UpdateDistrRecords(ctx sdk.Context, records ...types.DistrRecord) error {
	recordsMap := make(map[uint64]types.DistrRecord)
	totalWeight := sdk.NewInt(0)

	for _, existingRecord := range k.GetDistrInfo(ctx).Records {
		recordsMap[existingRecord.GaugeId] = existingRecord
		totalWeight = totalWeight.Add(existingRecord.Weight)
	}

	lastGaugeID := uint64(0)
	// TODO: Delete gaugeIdFlags
	gaugeIdFlags := make(map[uint64]bool)

	for _, record := range records {
		if gaugeIdFlags[record.GaugeId] {
			return sdkerrors.Wrapf(
				types.ErrDistrRecordRegisteredGauge,
				"Gauge ID #%d has duplications.",
				record.GaugeId,
			)
		}

		// Ensure records are sorted because ~AESTHETIC~
		if record.GaugeId < lastGaugeID {
			return sdkerrors.Wrapf(
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
				return sdkerrors.Wrapf(types.ErrDistrRecordRegisteredGauge,
					"Gauge ID #%d is not perpetual.",
					record.GaugeId)
			}
		}

		gaugeIdFlags[record.GaugeId] = true

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
		if !val.Weight.Equal(sdk.ZeroInt()) {
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
