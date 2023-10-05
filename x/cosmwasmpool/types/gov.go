package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeUploadCosmWasmPoolCodeAndWhiteList = "UploadCosmWasmPoolCodeAndWhiteListProposal"
	ProposalTypeMigratePoolContractsProposal       = "MigratePoolContractsProposal"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUploadCosmWasmPoolCodeAndWhiteList)
	govtypes.RegisterProposalTypeCodec(&UploadCosmWasmPoolCodeAndWhiteListProposal{}, "osmosis/UploadCosmWasmPoolCodeAndWhiteListProposal")
	govtypes.RegisterProposalType(ProposalTypeMigratePoolContractsProposal)
	govtypes.RegisterProposalTypeCodec(&MigratePoolContractsProposal{}, "osmosis/MigratePoolContractsProposal")
}

var (
	_ govtypes.Content = &UploadCosmWasmPoolCodeAndWhiteListProposal{}
	_ govtypes.Content = &MigratePoolContractsProposal{}
)

// NewUploadCosmWasmPoolCodeAndWhiteListProposal returns a new instance of an upload cosmwasm pool code and whitelist proposal struct.
func NewUploadCosmWasmPoolCodeAndWhiteListProposal(title, description string, wasmByteCode []byte) govtypes.Content {
	return &UploadCosmWasmPoolCodeAndWhiteListProposal{
		Title:        title,
		Description:  description,
		WASMByteCode: wasmByteCode,
	}
}

func (p *UploadCosmWasmPoolCodeAndWhiteListProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *UploadCosmWasmPoolCodeAndWhiteListProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *UploadCosmWasmPoolCodeAndWhiteListProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *UploadCosmWasmPoolCodeAndWhiteListProposal) ProposalType() string {
	return ProposalTypeUploadCosmWasmPoolCodeAndWhiteList
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *UploadCosmWasmPoolCodeAndWhiteListProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	if len(p.WASMByteCode) == 0 {
		return fmt.Errorf("wasm byte code cannot be nil")
	}

	return nil
}

// String returns a string containing the pool incentives proposal.
func (p UploadCosmWasmPoolCodeAndWhiteListProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Upload CosmWasm Pool Code and WhiteList Proposal:
Title:       %s
Description: %s
`, p.Title, p.Description))
	return b.String()
}

// NewMigratePoolContractsProposal returns a new instance of a contact code migration proposal.
func NewMigratePoolContractsProposal(title, description string, poolCodeIds []uint64, newCodeId uint64, wasmByteCode []byte) govtypes.Content {
	return &MigratePoolContractsProposal{
		Title:        title,
		Description:  description,
		PoolIds:      []uint64{},
		NewCodeId:    newCodeId,
		WASMByteCode: wasmByteCode,
	}
}

func (p *MigratePoolContractsProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *MigratePoolContractsProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *MigratePoolContractsProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *MigratePoolContractsProposal) ProposalType() string {
	return ProposalTypeMigratePoolContractsProposal
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *MigratePoolContractsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	if err := ValidateMigrationProposalConfiguration(p.PoolIds, p.NewCodeId, p.WASMByteCode); err != nil {
		return err
	}

	return nil
}

// String returns a string containing the pool incentives proposal.
func (p MigratePoolContractsProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Migrate CosmWasm Pool Code and WhiteList Proposal:
Title:       %s
Description: %s
PoolIds: %v
NewCodeId:   %d
Upload Wasm Code Given: %t
`, p.Title, p.Description, p.PoolIds, p.NewCodeId, len(p.WASMByteCode) > 0))
	return b.String()
}

// ValidateMigrationProposalConfiguration validates the migration proposal configuration.
// It has two options to perform the migration.
//
// 1. If the codeID is non-zero, it will migrate the pool contracts to a given codeID assuming that it has already
// been uploaded. uploadByteCode must be empty in such a case. Fails if codeID does not exist.
// Fails if uploadByteCode is not empty.
//
// 2. If the codeID is zero, it will upload the given uploadByteCode and use the new resulting code id to migrate
// the pool to. Errors if uploadByteCode is empty or invalid.
//
// For any of the options, it also validates that pool id list is not empty. Returns error if it is.
func ValidateMigrationProposalConfiguration(poolIds []uint64, newCodeId uint64, uploadByteCode []byte) error {
	if len(poolIds) == 0 {
		return ErrEmptyPoolIds
	}

	isNewCodeIdGiven := newCodeId != 0
	isUploadByteCodeGiven := len(uploadByteCode) != 0
	if !isNewCodeIdGiven && !isUploadByteCodeGiven {
		return ErrNoneOfCodeIdAndContractCodeSpecified
	}
	if isNewCodeIdGiven && isUploadByteCodeGiven {
		return ErrBothOfCodeIdAndContractCodeSpecified
	}
	return nil
}
