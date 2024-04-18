package types

var MigratedIncentiveAccumulatorPoolIDs = map[uint64]struct{}{
	1423: {},
	1213: {},
	1298: {},
	1297: {},
	1292: {},
	1431: {},
}

// MigratedSpreadFactorAccumulatorPoolIDsV25 is a map that defines pools to migrate to the latest scalingFactor for v25.
// These are ordered from the pool that has triggered the truncation alert the most to the least.
var MigratedSpreadFactorAccumulatorPoolIDsV25 = map[uint64]struct{}{
	// token0 usdc
	// token1 dai
	// liquidity 55.8k
	// https://app.osmosis.zone/pool/1260
	1260: {},
	// token0 eth
	// token1 usdc
	// liquidity 449k
	// https://app.osmosis.zone/pool/1264
	1264: {},
	// token0 osmo
	// token1 eth
	// liquidity 395k
	// https://app.osmosis.zone/pool/1477
	1477: {},
	// token0 osmo
	// token1 eth
	// liquidity 226k
	// https://app.osmosis.zone/pool/1134
	1134: {},
	// token0 usdt
	// token1 dai
	// liquidity 4.85k
	// https://app.osmosis.zone/pool/1261
	1261: {},
	// token0 eth
	// token1 osmo
	// liquidity 554k
	// https://app.osmosis.zone/pool/1281
	1281: {},
	// token0 stdydx
	// token1 dydx
	// liquidity 418k
	// https://app.osmosis.zone/pool/1423
	1423: {},
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
