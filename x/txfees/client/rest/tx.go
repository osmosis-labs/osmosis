package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
)

func ProposalUpdateFeeTokenProposalRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "update-fee-token",
		Handler:  emptyHandler(clientCtx),
	}
}

func emptyHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
