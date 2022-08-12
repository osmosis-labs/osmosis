package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v10/x/pool-incentives/types"
)

type UpdatePoolIncentivesRequest struct {
	BaseReq     rest.BaseReq        `json:"base_req" yaml:"base_req"`
	Title       string              `json:"title" yaml:"title"`
	Description string              `json:"description" yaml:"description"`
	Deposit     sdk.Coins           `json:"deposit" yaml:"deposit"`
	Records     []types.DistrRecord `json:"records" yaml:"records"`
}

// ProposalUpdatePoolIncentivesRESTHandler returns pool incentives update governance proposal handler.
func ProposalUpdatePoolIncentivesRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "update-pool-incentives",
		Handler:  newUpdatePoolIncentivesHandler(clientCtx),
	}
}

// newUpdatePoolIncentivesHandler creates a handler for pool incentives updates.
func newUpdatePoolIncentivesHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UpdatePoolIncentivesRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		content := types.NewUpdatePoolIncentivesProposal(req.Title, req.Description, req.Records)
		msg, err := govtypes.NewMsgSubmitProposal(content, req.Deposit, fromAddr)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

type ReplacePoolIncentivesRequest struct {
	BaseReq     rest.BaseReq        `json:"base_req" yaml:"base_req"`
	Title       string              `json:"title" yaml:"title"`
	Description string              `json:"description" yaml:"description"`
	Deposit     sdk.Coins           `json:"deposit" yaml:"deposit"`
	Records     []types.DistrRecord `json:"records" yaml:"records"`
}

func ProposalReplacePoolIncentivesRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "replace-pool-incentives",
		Handler:  newReplacePoolIncentivesHandler(clientCtx),
	}
}

// newReplacePoolIncentivesHandler creates a handler for pool incentives replacements.
func newReplacePoolIncentivesHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReplacePoolIncentivesRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		content := types.NewReplacePoolIncentivesProposal(req.Title, req.Description, req.Records)
		msg, err := govtypes.NewMsgSubmitProposal(content, req.Deposit, fromAddr)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}
