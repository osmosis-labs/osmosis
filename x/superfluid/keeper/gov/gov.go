package gov

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func HandleSetSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, ek types.EpochKeeper, p *types.SetSuperfluidAssetsProposal) error {
	for _, asset := range p.Assets {
		// Add check to ensure concentrated LP shares are formatted correctly
		if strings.HasPrefix(asset.Denom, cltypes.ConcentratedLiquidityTokenPrefix) {
			if asset.AssetType != types.SuperfluidAssetTypeConcentratedShare {
				return fmt.Errorf("concentrated LP share denom (%s) must have asset type %s", asset.Denom, types.SuperfluidAssetTypeConcentratedShare)
			}
		}
		if err := k.AddNewSuperfluidAsset(ctx, asset); err != nil {
			return err
		}
		events.EmitSetSuperfluidAssetEvent(ctx, asset.Denom, asset.AssetType)
	}
	return nil
}

func HandleRemoveSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, p *types.RemoveSuperfluidAssetsProposal) error {
	for _, denom := range p.SuperfluidAssetDenoms {
		asset, err := k.GetSuperfluidAsset(ctx, denom)
		if err != nil {
			return err
		}
		dummyAsset := types.SuperfluidAsset{}
		if asset.Equal(dummyAsset) {
			return fmt.Errorf("superfluid asset %s doesn't exist", denom)
		}
		k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
		events.EmitRemoveSuperfluidAsset(ctx, denom)
	}
	return nil
}

// HandleUnpoolWhiteListChange handles the unpool whitelist change proposal. It validates that every new pool id exists. Fails if not.
// If IsOverwrite flag is set, the whitelist is completely overridden. Otherwise, it is merged with pre-existing whitelisted pool ids.
// Any duplicates are removed and the pool ids are sorted prior to being written to state.
// Returns nil on success, error on failure.
func HandleUnpoolWhiteListChange(ctx sdk.Context, k keeper.Keeper, gammKeeper types.GammKeeper, p *types.UpdateUnpoolWhiteListProposal) error {
	allPoolIds := make([]uint64, 0, len(p.Ids))

	// if overwrite flag is not set, we merge the old white list with the
	// newly added pool ids.
	if !p.IsOverwrite {
		allPoolIds = append(allPoolIds, k.GetUnpoolAllowedPools(ctx)...)
	}

	for _, newId := range p.Ids {
		if newId == 0 {
			return errors.New("pool id 0 is not allowed. Pool ids start from 0")
		}

		if _, err := gammKeeper.GetPoolAndPoke(ctx, newId); err != nil {
			return fmt.Errorf("failed to get pool with id (%d), likely does not exist: %w", newId, err)
		}
		allPoolIds = append(allPoolIds, newId)
	}

	// Sort
	sort.Slice(allPoolIds, func(i, j int) bool {
		return allPoolIds[i] < allPoolIds[j]
	})

	// Remove duplicates, if any
	duplicatesRemovedIds := make([]uint64, 0, len(allPoolIds))
	for i, curId := range allPoolIds {
		if i < len(allPoolIds)-1 && curId == allPoolIds[i+1] {
			continue
		}
		duplicatesRemovedIds = append(duplicatesRemovedIds, curId)
	}

	k.SetUnpoolAllowedPools(ctx, duplicatesRemovedIds)
	return nil
}
