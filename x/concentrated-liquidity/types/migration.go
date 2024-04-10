package types

var MigratedSpreadFactorAccumulatorPoolIDs = map[uint64]struct{}{
	// token0 usdc
	// token1 dai
	// liquidity 55.8k
	1260: {},
	// token0 eth
	// token1 usdc
	// liquidity 449k
	1264: {},
	// token0 osmo
	// token1 eth
	// liquidity 395k
	1477: {},
	// token0 osmo
	// token1 eth
	// liquidity 226k
	1134: {},
	// token0 usdt
	// token1 dai
	// liquidity 4.85k
	1261: {},
	// token0 eth
	// token1 osmo
	// liquidity 554k
	1281: {},
}

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
