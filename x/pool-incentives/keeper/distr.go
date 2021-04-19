package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/pool-incentives/types"
)

// GetAllocatableAsset gets the balance of the `MintedDenom` from the `feeCollectorName` module account and returns coins according to the `AllocationRatio`
func (k Keeper) GetAllocatableAsset(ctx sdk.Context) sdk.Coin {
	params := k.GetParams(ctx)

	feeCollector := k.accountKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	asset := k.bankKeeper.GetBalance(ctx, feeCollector.GetAddress(), params.MintedDenom)

	return sdk.NewCoin(asset.Denom, asset.Amount.ToDec().Mul(params.AllocationRatio).TruncateInt())
}

// AllocateAsset allocates and distributes coin according a pot’s proportional weight that is recorded in the record
func (k Keeper) AllocateAsset(ctx sdk.Context, asset sdk.Coin) error {
	distrInfo := k.GetDistrInfo(ctx)

	if distrInfo.TotalWeight.GT(sdk.ZeroInt()) {
		assetAmountDec := asset.Amount.ToDec()
		totalWeightDec := distrInfo.TotalWeight.ToDec()
		for _, record := range distrInfo.Records {
			allocatingAmount := assetAmountDec.Mul(record.Weight.ToDec().Quo(totalWeightDec)).TruncateInt()
			coins := sdk.NewCoins(sdk.NewCoin(asset.Denom, allocatingAmount))

			err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, coins)
			if err != nil {
				return err
			}

			err = k.incentivesKeeper.AddToPotRewards(ctx, k.accountKeeper.GetModuleAddress(types.ModuleName), coins, record.PotId)
			if err != nil {
				return err
			}
		}
	} else {
		// If there are no records, put the asset to the community pool
		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, k.communityPoolName, sdk.Coins{asset})
		if err != nil {
			return err
		}

		feePool := k.distrKeeper.GetFeePool(ctx)
		feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(asset)...)
		k.distrKeeper.SetFeePool(ctx, feePool)
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

func (k Keeper) AddDistrRecords(ctx sdk.Context, records ...types.DistrRecord) error {
	distrInfo := k.GetDistrInfo(ctx)

	deltaWeight := sdk.NewInt(0)
	for _, record := range records {
		// Make sure that the pot exists.
		_, err := k.incentivesKeeper.GetPotByID(ctx, record.PotId)
		if err != nil {
			return err
		}

		deltaWeight = deltaWeight.Add(record.Weight)
	}

	distrInfo.TotalWeight = distrInfo.TotalWeight.Add(deltaWeight)
	distrInfo.Records = append(distrInfo.Records, records...)

	k.SetDistrInfo(ctx, distrInfo)

	return nil
}

func (k Keeper) EditDistrRecords(ctx sdk.Context, records ...types.EditPoolIncentivesProposal_DistrRecordWithIndex) error {
	distrInfo := k.GetDistrInfo(ctx)

	deltaWeight := sdk.NewInt(0)
	for _, recordWithIndex := range records {
		index := recordWithIndex.Index
		record := recordWithIndex.Record

		if index < 0 || uint64(len(distrInfo.Records)) <= index {
			return types.ErrDistrRecordInvalidIndex
		}

		priorRecord := distrInfo.Records[index]

		if priorRecord.PotId != record.PotId {
			return types.ErrDistrRecordMismatchedPotId
		}

		deltaWeight = deltaWeight.Add(record.Weight.Sub(priorRecord.Weight))

		distrInfo.Records[index] = record
	}

	distrInfo.TotalWeight = distrInfo.TotalWeight.Add(deltaWeight)

	k.SetDistrInfo(ctx, distrInfo)

	return nil
}

func (k Keeper) RemoveDistrRecords(ctx sdk.Context, indexes ...uint64) error {
	distrInfo := k.GetDistrInfo(ctx)

	for _, index := range indexes {
		if index < 0 || uint64(len(distrInfo.Records)) <= index {
			return types.ErrDistrRecordInvalidIndex
		}

		record := distrInfo.Records[index]
		distrInfo.TotalWeight = distrInfo.TotalWeight.Sub(record.Weight)
		distrInfo.Records = append(distrInfo.Records[0:index], distrInfo.Records[index+1:]...)
	}

	k.SetDistrInfo(ctx, distrInfo)

	return nil
}
