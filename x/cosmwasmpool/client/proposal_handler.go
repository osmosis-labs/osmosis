package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/client/rest"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
)

var (
	UploadCodeIdAndWhitelistProposalHandler = govclient.NewProposalHandler(cli.NewCmdUploadCodeIdAndWhitelistProposal, rest.ProposalTickSpacingDecreaseRESTHandler)
	MigratePoolContractsProposalHandler     = govclient.NewProposalHandler(cli.NewCmdMigratePoolContractsProposal, rest.ProposalCreateConcentratedLiquidityPoolHandler)
)

func UploadCodeIdAndWhitelistProposalRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "upload-code-id-and-whitelist",
		Handler:  emptyHandler(clientCtx),
	}
}

func MigratePoolContractsProposalRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "migrate-cw-pool-contracts",
		Handler:  emptyHandler(clientCtx),
	}
}

func emptyHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
