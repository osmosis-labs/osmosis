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
	// v16 upgrade specific canonical OSMO/DAI pool id.
	// It is expected to create a concentrated liquidity pool
	// associated with this balancer pool in the upgrade handler.
	// This is meant to be removed post-v16.
	DaiOsmoPoolIdv16 uint64
	// A pool created via CLI before starting an
	// upgrade.
	PreUpgradePoolId uint64

	PreUpgradeStableSwapPoolId uint64

	StrideMigrateWallet = "stride-migration"

	LockupWallet = "lockup-wallet"

	LockupWalletSuperfluid = "lockup-wallet-superfluid"

	StableswapWallet = "stableswap-wallet"
)
