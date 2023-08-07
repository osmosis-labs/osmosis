package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
)

func ProposalReplaceMigrationRecordsRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "replace-migrations-records",
		Handler:  emptyHandler(clientCtx),
	}
}

func ProposalUpdateMigrationRecordsRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "update-migrations-records",
		Handler:  emptyHandler(clientCtx),
	}
}

func ProposalCreateConcentratedLiquidityPoolAndLinkToCFMMHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "create-cl-pool-and-cfmm-link",
		Handler:  emptyHandler(clientCtx),
	}
}

func ProposalSetScalingFactorController(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "set-scaling-factor-controller",
		Handler:  emptyHandler(clientCtx),
	}
}

func emptyHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
