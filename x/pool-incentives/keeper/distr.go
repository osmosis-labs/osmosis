package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/x/pool-incentives/types"
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

// AllocateAsset allocates and distributes coin according a pot’s proportional weight that is recorded in the record
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
			logger.Info(fmt.Sprintf("allocating amount for (%d, %s) record is not positive", record.PotId, record.Weight.String()))
			continue
		}

		if record.PotId == 0 { // fund community pool if potId is zero
			k.FundCommunityPoolFromModule(ctx, sdk.NewCoin(asset.Denom, allocatingAmount))
			continue
		}

		coins := sdk.NewCoins(sdk.NewCoin(asset.Denom, allocatingAmount))
		err := k.incentivesKeeper.AddToPotRewards(ctx, k.accountKeeper.GetModuleAddress(types.ModuleName), coins, record.PotId)
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

// indexOfDistrRecordByPotId returns the index of the record for the specific pot id.
// If there is no record matched to the pot id, return -1.
func (k Keeper) indexOfDistrRecordByPotId(ctx sdk.Context, potId uint64) int {
	distrInfo := k.GetDistrInfo(ctx)
	records := distrInfo.Records
	for i, record := range records {
		if record.PotId == potId {
			return i
		}
	}
	return -1
}

func (k Keeper) UpdateDistrRecords(ctx sdk.Context, records ...types.DistrRecord) error {
	distrInfo := k.GetDistrInfo(ctx)

	potIdFlags := make(map[uint64]bool)

	totalWeight := sdk.NewInt(0)
	for _, record := range records {
		if potIdFlags[record.PotId] {
			return sdkerrors.Wrapf(
				types.ErrDistrRecordRegisteredPot,
				"Pot ID #%d has duplications.",
				record.PotId,
			)
		}

		// unless PotID is 0 for the community pool, don't allow distribution records for pots that don't exist
		if record.PotId != 0 {
			_, err := k.incentivesKeeper.GetPotByID(ctx, record.PotId)
			if err != nil {
				return err
			}
		}

		potIdFlags[record.PotId] = true
		totalWeight = totalWeight.Add(record.Weight)
	}

	distrInfo.Records = records
	distrInfo.TotalWeight = totalWeight

	k.SetDistrInfo(ctx, distrInfo)
	return nil
}
