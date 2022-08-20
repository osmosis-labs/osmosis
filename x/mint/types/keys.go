package types

var (
	// MinterKey is the key to use for the keeper store at which
	// the Minter and its EpochProvisions are stored.
	MinterKey = []byte{0x00}

	// LastReductionEpochKey is the key to use for the keeper store
	// for storing the last epoch at which reduction occurred.
	LastReductionEpochKey = []byte{0x03}

	// TruncatedInflationDeltaKey represents key for the the delta of minted
	// inflation coins that have not been distributed yet due to truncations..
	// Truncations are stemming from the interfaces of the core modules such as
	// bank and distribution that operate on integers. Decimals allow
	// for much higher precision. As a result, by storing decimal delta
	// we can avoid truncation discrepancies and be in-line with the
	// projected total supply of OSMO.
	TruncatedInflationDeltaKey = []byte{0x04}

	// TruncatedDeveloperVestingDelta represents the delta of developer
	// vesting rewards that has not been distributed yet due to truncations.
	// Truncations are stemming from the interfaces of the core modules such as
	// bank and distribution that operate on integers. As a result, by
	// storing decimal delta we can avoid truncation discrepancies and be
	// in-line with the projected total supply of OSMO.
	TruncatedDeveloperVestingDeltaKey = []byte{0x05}
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
