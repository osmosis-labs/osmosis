package types

var MigratedIncentiveAccumulatorPoolIDs = map[uint64]struct{}{
	1423: {},
	1213: {},
	1298: {},
	1297: {},
	1292: {},
	1431: {},
}

// MigratedIncentiveAccumulatorPoolIDsV24 is a map that defines pools to migrate to the latest scalingFactor for v24.
var MigratedIncentiveAccumulatorPoolIDsV24 = map[uint64]struct{}{
	// token0 eth
	// token1 uosmo
	// liquidity 272k
	1281: {},
	// token0 usdt
	// token1 dia
	// liquidity 500
	1276: {},
	// token0 usdc
	// token1 dai
	// liquidity 1.55k
	1275: {},
	// token0 usdc
	// token1 dai
	// liquidity 52.9k
	1260: {},
}
