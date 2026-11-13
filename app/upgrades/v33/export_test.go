package v33

// CollectHolders exposes collectHolders so tests can check which holders the migration visits.
var CollectHolders = collectHolders

// ParseAddresses and WBTCUnindexedHolders expose the embedded unindexed holder list.
var (
	ParseAddresses       = parseAddresses
	WBTCUnindexedHolders = wbtcUnindexedHolders
)
