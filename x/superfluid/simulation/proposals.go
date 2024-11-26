package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

const (
	OpWeightSetSuperfluidAssetsProposal    = "op_weight_set_superfluid_assets_proposal"
	OpWeightRemoveSuperfluidAssetsProposal = "op_weight_remove_superfluid_assets_proposal"
)

// ProposalContents defines the module weighted proposals' contents.
//
//nolint:staticcheck
func ProposalContents(k keeper.Keeper, gk types.GammKeeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSetSuperfluidAssetsProposal,
			DefaultWeightSetSuperfluidAssetsProposal,
			SimulateSetSuperfluidAssetsProposal(k, gk),
		),
	}
}

// SimulateSetSuperfluidAssetsProposal generates random superfluid asset set proposal content.
//
//nolint:staticcheck
func SimulateSetSuperfluidAssetsProposal(k keeper.Keeper, gk types.GammKeeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		pools, err := gk.GetPoolsAndPoke(ctx)
		if err != nil {
			return nil
		}

		if len(pools) == 0 {
			return nil
		}

		poolIndex := r.Intn(len(pools))
		pool := pools[poolIndex]

		return &types.SetSuperfluidAssetsProposal{
			Title:       "set superfluid assets",
			Description: "set superfluid assets description",
			Assets: []types.SuperfluidAsset{
				{
					Denom:     gammtypes.GetPoolShareDenom(pool.GetId()),
					AssetType: types.SuperfluidAssetTypeLPShare,
				},
			},
		}
	}
}

// SimulateRemoveSuperfluidAssetsProposal generates random superfluid asset removal proposal content.
//
//nolint:staticcheck
func SimulateRemoveSuperfluidAssetsProposal(k keeper.Keeper, gk types.GammKeeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		assets := k.GetAllSuperfluidAssets(ctx)

		if len(assets) == 0 {
			return nil
		}

		assetIndex := r.Intn(len(assets))
		asset := assets[assetIndex]

		return &types.RemoveSuperfluidAssetsProposal{
			Title:                 "remove superfluid assets",
			Description:           "remove superfluid assets description",
			SuperfluidAssetDenoms: []string{asset.Denom},
		}
	}
}
