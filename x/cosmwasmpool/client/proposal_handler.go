package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/osmosis-labs/osmosis/v19/x/cosmwasmpool/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	UploadCodeIdAndWhitelistProposalHandler = govclient.NewProposalHandler(cli.NewCmdUploadCodeIdAndWhitelistProposal, UploadCodeIdAndWhitelistProposalRESTHandler)
	MigratePoolContractsProposalHandler     = govclient.NewProposalHandler(cli.NewCmdMigratePoolContractsProposal, MigratePoolContractsProposalRESTHandler)
)

func emptyHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
