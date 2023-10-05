package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/osmosis-labs/osmosis/v19/x/cosmwasmpool/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
)

var (
	UploadCodeIdAndWhitelistProposalHandler = govclient.NewProposalHandler(cli.NewCmdUploadCodeIdAndWhitelistProposal, UploadCodeIdAndWhitelistProposalRESTHandler)
	MigratePoolContractsProposalHandler     = govclient.NewProposalHandler(cli.NewCmdMigratePoolContractsProposal, MigratePoolContractsProposalRESTHandler)
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
