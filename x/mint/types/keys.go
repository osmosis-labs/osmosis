package types

var (
	// MinterKey is the key to use for the keeper store at which
	// the Minter and its EpochProvisions are stored.
	MinterKey = []byte{0x00}

	// LastReductionEpochKey is the key to use for the keeper store
	// for storing the last epoch at which reduction occurred.
	LastReductionEpochKey = []byte{0x03}

	// LastMintedTotalAmount is the key to use for the keeper store
	// for storing the last minted accumulator value.
	// It represents the total amount of tokens minted since the
	// chain launched.
	LastMintedTotalAmount = []byte{0x04}
)

const (
	// ModuleName is the module name.
	ModuleName = "mint"
	// DeveloperVestingModuleAcctName is the module acct name for developer vesting.
	DeveloperVestingModuleAcctName = "developer_vesting_unvested"

	// StoreKey is the default store key for mint.
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the minting store.
	QuerierRoute = StoreKey

	// QueryParameters is an endpoint path for querying mint parameters.
	QueryParameters = "parameters"

	// QueryEpochProvisions is an endpoint path for querying mint epoch provisions.
	QueryEpochProvisions = "epoch_provisions"
)
