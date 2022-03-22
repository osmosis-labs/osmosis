package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
)

func ProposalSetSuperfluidAssetsRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "set-superfluid-assets",
		Handler:  newSetSuperfluidAssetsHandler(clientCtx),
	}
}

func newSetSuperfluidAssetsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func ProposalRemoveSuperfluidAssetsRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "remove-superfluid-assets",
		Handler:  newRemoveSuperfluidAssetsHandler(clientCtx),
	}
}

func newRemoveSuperfluidAssetsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
