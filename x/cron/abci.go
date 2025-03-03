package cron

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v29/x/cron/keeper"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	crons := k.GetCronJobs(ctx)
	for _, cron := range crons {
		if cron.EnableCron {
			for _, job := range cron.MsgContractCron {
				err := k.SudoContractCall(ctx, job.ContractAddress, []byte(job.JsonMsg))
				if err != nil {
					ctx.Logger().Error("Cronjob failed for: ", cron.Name, " contract: ", job.ContractAddress)
				} else {
					ctx.Logger().Info("Cronjob success for: ", cron.Name, " contract: ", job.ContractAddress)
				}
			}
		}
	}
}
