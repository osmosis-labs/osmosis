package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/c-osmosis/osmosis/x/pool-incentives/types"
)

type AddPoolIncentivesRequest struct {
	BaseReq     rest.BaseReq        `json:"base_req" yaml:"base_req"`
	Title       string              `json:"title" yaml:"title"`
	Description string              `json:"description" yaml:"description"`
	Deposit     sdk.Coins           `json:"deposit" yaml:"deposit"`
	Records     []types.DistrRecord `json:"records" yaml:"records"`
}

type EditPoolIncentivesRequest struct {
	BaseReq     rest.BaseReq                                            `json:"base_req" yaml:"base_req"`
	Title       string                                                  `json:"title" yaml:"title"`
	Description string                                                  `json:"description" yaml:"description"`
	Deposit     sdk.Coins                                               `json:"deposit" yaml:"deposit"`
	Records     []types.EditPoolIncentivesProposal_DistrRecordWithIndex `json:"records" yaml:"records"`
}

type RemovePoolIncentivesRequest struct {
	BaseReq     rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title       string       `json:"title" yaml:"title"`
	Description string       `json:"description" yaml:"description"`
	Deposit     sdk.Coins    `json:"deposit" yaml:"deposit"`
	Indexes     []uint64     `json:"indexes" yaml:"indexes"`
}

func ProposalAddPoolIncentivesRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "poolyield",
		Handler:  newAddPoolIncentivesHandler(clientCtx),
	}
}

func ProposalEditPoolIncentivesRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "poolyield",
		Handler:  newEditPoolIncentivesHandler(clientCtx),
	}
}

func ProposalRemovePoolIncentivesRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "poolyield",
		Handler:  newRemovePoolIncentivesHandler(clientCtx),
	}
}

func newAddPoolIncentivesHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AddPoolIncentivesRequest

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

		content := types.NewAddPoolIncentivesProposal(req.Title, req.Description, req.Records)
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

func newEditPoolIncentivesHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req EditPoolIncentivesRequest

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

		content := types.NewEditPoolIncentivesProposal(req.Title, req.Description, req.Records)
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

func newRemovePoolIncentivesHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RemovePoolIncentivesRequest

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

		content := types.NewRemovePoolIncentivesProposal(req.Title, req.Description, req.Indexes)
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
