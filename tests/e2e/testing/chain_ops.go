package e2eTesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"
)

// ExecuteGovProposal submits a new proposal and votes for it.
func (chain *TestChain) ExecuteGovProposal(proposerAcc Account, expPass bool, proposals []sdk.Msg, title string, summary string, metadata string) {
	t := chain.t

	// Get params
	k := chain.app.GovKeeper
	govParams, err := k.Params.Get(chain.GetContext())
	require.NoError(t, err)
	depositCoin := govParams.MinDeposit
	votingDur := govParams.VotingPeriod

	// Submit proposal with min deposit to start the voting
	msg, err := govTypes.NewMsgSubmitProposal(proposals, depositCoin, proposerAcc.Address.String(), metadata, title, summary, false)
	require.NoError(t, err)

	_, res, _, err := chain.SendMsgs(proposerAcc, true, []sdk.Msg{msg})
	require.NoError(t, err)
	txRes := chain.ParseSDKResultData(res)
	require.Len(t, txRes.MsgResponses, 1)

	var resp govTypes.MsgSubmitProposalResponse
	require.NoError(t, resp.Unmarshal(txRes.MsgResponses[0].Value))
	proposalID := resp.ProposalId

	// Vote with all validators (delegators)
	for i := 0; i < len(chain.valSet.Validators); i++ {
		delegatorAcc := chain.GetAccount(i)

		msg := govTypes.NewMsgVote(delegatorAcc.Address, proposalID, govTypes.OptionYes, "metadata")
		_, _, _, err = chain.SendMsgs(proposerAcc, true, []sdk.Msg{msg})
		require.NoError(t, err)
	}

	// Wait for voting to end
	chain.NextBlock(*votingDur)
	chain.NextBlock(0) // for the Gov EndBlocker to work

	// Check if proposal was passed
	proposal, err := k.Proposals.Get(chain.GetContext(), proposalID)
	require.NoError(t, err)

	if expPass {
		require.Equal(t, govTypes.StatusPassed.String(), proposal.Status.String())
	} else {
		require.NotEqual(t, govTypes.StatusPassed.String(), proposal.Status.String())
	}
}
