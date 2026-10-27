package v33

// CollectHolders exposes collectHolders so tests can exercise the holder cap without creating
// MaxWBTCHoldersToMigrate accounts.
var CollectHolders = collectHolders

// ParseAddresses and WBTCUnindexedHolders expose the embedded unindexed holder list.
var (
	ParseAddresses       = parseAddresses
	WBTCUnindexedHolders = wbtcUnindexedHolders
)
