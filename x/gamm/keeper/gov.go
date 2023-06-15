package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

func (k Keeper) HandleReplaceMigrationRecordsProposal(ctx sdk.Context, p *types.ReplaceMigrationRecordsProposal) error {
	return k.ReplaceMigrationRecords(ctx, p.Records)
}

func (k Keeper) HandleUpdateMigrationRecordsProposal(ctx sdk.Context, p *types.UpdateMigrationRecordsProposal) error {
	return k.UpdateMigrationRecords(ctx, p.Records)
}

// HandleLinkBalancerPoolWithCLPoolProposal creates link between an existing CL pool and Balancer Pool.
// TODO: If the CL pool doesnot exist create one
func (k Keeper) HandleLinkBalancerPoolWithCLPoolProposal(ctx sdk.Context, p *types.LinkBalancerPoolWithCLPoolProposal) error {
	// both cfmm and cl poolId exists with same denom
	err := k.validateRecords(ctx, p.Records)
	if err != nil {
		return err
	}

	return k.OverwriteMigrationRecordsAndRedirectDistrRecords(ctx, types.MigrationRecords{
		BalancerToConcentratedPoolLinks: p.Records,
	})
}

// func (k Keeper) BalancerExistCLPoolDoesnotExist(ctx sdk.Context, cfmmPoolIdToLinkWith uint64) (uint64, error) {
// 	cfmmPool, err := k.GetCFMMPool(ctx, cfmmPoolIdToLinkWith)
// 	if err != nil {
// 		return 0, err
// 	}

// 	poolLiquidity := cfmmPool.GetTotalPoolLiquidity(ctx)
// 	if len(poolLiquidity) != 2 {
// 		return 0, nil
// 	}

// 	poolmanagerModuleAccAddress := k.accountKeeper.GetModuleAccount(ctx, poolmanagertypes.ModuleName).GetAddress()
// 	createPoolMsg := clmodel.NewMsgCreateConcentratedPool(poolmanagerModuleAccAddress, poolLiquidity[0].Denom, poolLiquidity[1].Denom, 1, cfmmPool.GetSpreadFactor(ctx))
// 	clPool, err := k.poolManager.CreateConcentratedPoolAsPoolManager(ctx, createPoolMsg)
// 	if err != nil {
// 		return 0, err
// 	}

// 	return clPool.GetId(), nil
// }
