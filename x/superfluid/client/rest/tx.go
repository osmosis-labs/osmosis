package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
)

func ProposalSetSuperfluidAssetsRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "set-superfluid-assets",
		Handler:  emptyHandler(clientCtx),
	}
}

func ProposalRemoveSuperfluidAssetsRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "remove-superfluid-assets",
		Handler:  emptyHandler(clientCtx),
	}
}

func ProposalUpdateUnpoolWhitelistProposal(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "update-unpool-whitelist",
		Handler:  emptyHandler(clientCtx),
	}
}

func emptyHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
