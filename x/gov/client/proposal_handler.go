package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// function to create the rest handler
type RESTHandlerFn = govclient.RESTHandlerFn

// function to create the cli handler
type CLIHandlerFn = govclient.CLIHandlerFn

// The combined type for a proposal handler for both cli and rest
type ProposalHandler = govclient.ProposalHandler

// NewProposalHandler creates a new ProposalHandler object
func NewProposalHandler(cliHandler CLIHandlerFn, restHandler RESTHandlerFn) ProposalHandler {
	return ProposalHandler{
		CLIHandler:  cliHandler,
		RESTHandler: restHandler,
	}
}
