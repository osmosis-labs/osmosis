package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeSetSwapFee = "SetSwapFee"
	ProposalTypeSetExitFee = "SetExitFee"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeSetSwapFee)
	govtypes.RegisterProposalTypeCodec(&SetSwapFeeProposal{}, "osmosis/SetSwapFeeProposal")
	govtypes.RegisterProposalType(ProposalTypeSetExitFee)
	govtypes.RegisterProposalTypeCodec(&SetExitFeeProposal{}, "osmosis/SetExitFeeProposal")
}

var (
	_ govtypes.Content = &SetSwapFeeProposal{}
	_ govtypes.Content = &SetExitFeeProposal{}
)

func NewSetSwapFeeProposal(title, description string, poolId uint64, newSwapFee sdk.Dec) govtypes.Content {
	content := SetSwapFeeContent{
		PoolId:  poolId,
		SwapFee: newSwapFee,
	}

	return &SetSwapFeeProposal{
		Title:       title,
		Description: description,
		Content:     content,
	}
}

func (p *SetSwapFeeProposal) GetTitle() string { return p.Title }

func (p *SetSwapFeeProposal) GetDescription() string { return p.Description }

func (p *SetSwapFeeProposal) ProposalRoute() string { return RouterKey }

func (p *SetSwapFeeProposal) ProposalType() string {
	return ProposalTypeSetSwapFee
}

func (p *SetSwapFeeProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if p.Content.PoolId <= 0 {
		return ErrInvalidPoolId
	}
	if p.Content.SwapFee.LT(sdk.NewDec(0)) {
		return ErrNegativeSwapFee
	}
	return nil
}

func (p SetSwapFeeProposal) String() string {
	return fmt.Sprintf(`Set Swap Fee Proposal:
	Title:       %s
	Description: %s
	Content:     %+v
  `, p.Title, p.Description, p.Content)
}

func NewSetExitFeeProposal(title, description string, poolId uint64, newExitFee sdk.Dec) govtypes.Content {
	content := SetExitFeeContent{
		PoolId:  poolId,
		ExitFee: newExitFee,
	}
	return &SetExitFeeProposal{
		Title:       title,
		Description: description,
		Content:     content,
	}
}

func (p *SetExitFeeProposal) GetTitle() string { return p.Title }

func (p *SetExitFeeProposal) GetDescription() string { return p.Description }

func (p *SetExitFeeProposal) ProposalRoute() string { return RouterKey }

func (p *SetExitFeeProposal) ProposalType() string {
	return ProposalTypeSetExitFee
}

func (p *SetExitFeeProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if p.Content.PoolId <= 0 {
		return ErrInvalidPoolId
	}
	if p.Content.ExitFee.LT(sdk.NewDec(0)) {
		return ErrNegativeExitFee
	}
	return nil
}

func (p SetExitFeeProposal) String() string {
	return fmt.Sprintf(`Set Exit Fee Proposal:
	Title:       %s
	Description: %s
	Content:     %+v
  `, p.Title, p.Description, p.Content)
}
