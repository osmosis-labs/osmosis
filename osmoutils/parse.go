package osmoutils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/spf13/pflag"
)

type Proposal struct {
	Title       string
	Description string
	Deposit     string
}

var ProposalFlags = []string{
	cli.FlagTitle,
	cli.FlagDescription,
	cli.FlagDeposit,
}

func (p Proposal) validate() error {
	if p.Title == "" {
		return fmt.Errorf("proposal title is required")
	}

	if p.Description == "" {
		return fmt.Errorf("proposal description is required")
	}
	return nil
}

func ParseProposalFlags(fs *pflag.FlagSet) (*Proposal, error) {
	proposal := &Proposal{}
	proposalFile, _ := fs.GetString(cli.FlagProposal)

	if proposalFile == "" {
		proposal.Title, _ = fs.GetString(cli.FlagTitle)
		proposal.Description, _ = fs.GetString(cli.FlagDescription)
		proposal.Deposit, _ = fs.GetString(cli.FlagDeposit)
		if err := proposal.validate(); err != nil {
			return nil, err
		}

		return proposal, nil
	}

	for _, flag := range ProposalFlags {
		if v, _ := fs.GetString(flag); v != "" {
			return nil, fmt.Errorf("--%s flag provided alongside --proposal, which is a noop", flag)
		}
	}

	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, proposal)
	if err != nil {
		return nil, err
	}

	if err := proposal.validate(); err != nil {
		return nil, err
	}

	return proposal, nil
}
