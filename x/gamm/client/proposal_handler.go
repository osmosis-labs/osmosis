package client

import (
	"net/http"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/client/cli"
	"github.com/cosmos/cosmos-sdk/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	SetSwapFeeProposalHandler    = govclient.NewProposalHandler(cli.NewSetSwapFeeProposalCmd, ProposalSetSwapFeeRESTHandler)
	SetExitFeeProposalHandler 	 = govclient.NewProposalHandler(cli.NewSetExitFeeProposalCmd, ProposalSetExitFeeRESTHandler)
)

func ProposalSetSwapFeeRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "set-swap-fee",
		Handler:  newSetSwapFeeHandler(clientCtx),
	}
}

func newSetSwapFeeHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func ProposalSetExitFeeRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "set-exit-fee",
		Handler:  newSetSwapFeeHandler(clientCtx),
	}
}

func newSetExitFeeHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
