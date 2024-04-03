package types

var MigratedIncentiveAccumulatorPoolIDs = map[uint64]struct{}{
	1423: {},
	1213: {},
	1298: {},
	1297: {},
	1292: {},
	1431: {},
}

// FinalIncentiveAccumulatorPoolIDsToMigrate is a map that defines all pools to migrate to the latest scalingFactor.
// We store the latest pool to use the scalingFactor which is pool 1496, any pool created after this uses the
// new scalingFactor, so we only track pools that need to be migrated in this map.
var FinalIncentiveAccumulatorPoolIDsToMigrate = map[uint64]struct{}{
	// token0 ibc/4ABBEF4C8926DDDB320AE5188CFD63267ABBCEFC0583E4AE05D6E5AA2401DDAB
	// token1 ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7
	1276: {},
	// token0 ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4
	// token1 ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7
	1275: {},
}
