package types

// MinterKey is the key to use for the keeper store at which
// the Minter and its EpochProvisions are stored.
var MinterKey = []byte{0x00}

// LastReductionEpochKey is the key to use for the keeper store
// for storing the last epoch at which reduction occurred.
var LastReductionEpochKey = []byte{0x03}

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
