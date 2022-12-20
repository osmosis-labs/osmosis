package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"

	"github.com/osmosis-labs/osmosis/v13/x/protorev/client/cli"
)

var (
	SetProtoRevAdminAccountProposalHandler = govclient.NewProposalHandler(cli.CmdSetProtoRevAdminAccountProposal, ProposalSetProtoRevAdminAccountRESTHandler)
	SetProtoRevEnabledProposalHandler      = govclient.NewProposalHandler(cli.CmdSetProtoRevEnabledProposal, ProposalSetProtoRevEnabledRESTHandler)
)

func ProposalSetProtoRevAdminAccountRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "set-protorev-admin-account-proposal",
		Handler:  emptyHandler(clientCtx),
	}
}

func ProposalSetProtoRevEnabledRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "set-protorev-enabled-proposal",
		Handler:  emptyHandler(clientCtx),
	}
}

func emptyHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
