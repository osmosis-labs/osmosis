package config

import govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

const (
	// if not skipping upgrade, how many blocks we allow for fork to run pre upgrade state creation
	ForkHeightPreUpgradeOffset int64 = 60
	// estimated number of blocks it takes to submit for a proposal
	PropSubmitBlocks float32 = 10
	// estimated number of blocks it takes to deposit for a proposal
	PropDepositBlocks float32 = 10
	// number of blocks it takes to vote for a single validator to vote for a proposal
	PropVoteBlocks float32 = 1.2
	// number of blocks used as a calculation buffer
	PropBufferBlocks float32 = 6
	// max retries for json unmarshalling
	MaxRetries = 60
)

var (
	// Minimum deposit value for a proposal to enter a voting period.
	MinDepositValue = govtypes.DefaultMinDepositTokens.Int64()
	// Minimum expedited deposit value for a proposal to enter a voting period.
	MinExpeditedDepositValue = govtypes.DefaultMinExpeditedDepositTokens.Int64()
	// Minimum deposit value for proposal to be submitted.
	InitialMinDeposit = MinDepositValue / 4
	// Minimum expedited deposit value for proposal to be submitted.
	InitialMinExpeditedDeposit = MinExpeditedDepositValue / 4
	// The first id of a pool create via CLI before starting an
	// upgrade.
	// Note: that we create a pool with id 1 via genesis
	// in the initialization package. As a result, the first
	// pre-upgrade should have id 2.
	// This value gets mutated during the pre-upgrade pool
	// creation in case more pools are added to genesis
	// in the future
	PreUpgradePoolId uint64 = 2

	StrideMigrateWallet = "stride-migration"
)
