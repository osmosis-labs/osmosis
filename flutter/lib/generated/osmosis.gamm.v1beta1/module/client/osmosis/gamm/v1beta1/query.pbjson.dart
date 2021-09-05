///
//  Generated code. Do not modify.
//  source: osmosis/gamm/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const QueryPoolRequest$json = const {
  '1': 'QueryPoolRequest',
  '2': const [
    const {'1': 'poolId', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
  ],
};

const QueryPoolResponse$json = const {
  '1': 'QueryPoolResponse',
  '2': const [
    const {'1': 'pool', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'pool'},
  ],
};

const QueryPoolsRequest$json = const {
  '1': 'QueryPoolsRequest',
  '2': const [
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryPoolsResponse$json = const {
  '1': 'QueryPoolsResponse',
  '2': const [
    const {'1': 'pools', '3': 1, '4': 3, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'pools'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryNumPoolsRequest$json = const {
  '1': 'QueryNumPoolsRequest',
};

const QueryNumPoolsResponse$json = const {
  '1': 'QueryNumPoolsResponse',
  '2': const [
    const {'1': 'numPools', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'numPools'},
  ],
};

const QueryPoolParamsRequest$json = const {
  '1': 'QueryPoolParamsRequest',
  '2': const [
    const {'1': 'poolId', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
  ],
};

const QueryPoolParamsResponse$json = const {
  '1': 'QueryPoolParamsResponse',
  '2': const [
    const {'1': 'params', '3': 1, '4': 1, '5': 11, '6': '.osmosis.gamm.v1beta1.PoolParams', '8': const {}, '10': 'params'},
  ],
};

const QueryTotalSharesRequest$json = const {
  '1': 'QueryTotalSharesRequest',
  '2': const [
    const {'1': 'poolId', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
  ],
};

const QueryTotalSharesResponse$json = const {
  '1': 'QueryTotalSharesResponse',
  '2': const [
    const {'1': 'totalShares', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'totalShares'},
  ],
};

const QueryPoolAssetsRequest$json = const {
  '1': 'QueryPoolAssetsRequest',
  '2': const [
    const {'1': 'poolId', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
  ],
};

const QueryPoolAssetsResponse$json = const {
  '1': 'QueryPoolAssetsResponse',
  '2': const [
    const {'1': 'poolAssets', '3': 1, '4': 3, '5': 11, '6': '.osmosis.gamm.v1beta1.PoolAsset', '8': const {}, '10': 'poolAssets'},
  ],
};

const QuerySpotPriceRequest$json = const {
  '1': 'QuerySpotPriceRequest',
  '2': const [
    const {'1': 'poolId', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'tokenInDenom', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'tokenInDenom'},
    const {'1': 'tokenOutDenom', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'tokenOutDenom'},
    const {'1': 'withSwapFee', '3': 4, '4': 1, '5': 8, '8': const {}, '10': 'withSwapFee'},
  ],
};

const QuerySpotPriceResponse$json = const {
  '1': 'QuerySpotPriceResponse',
  '2': const [
    const {'1': 'spotPrice', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'spotPrice'},
  ],
};

const QuerySwapExactAmountInRequest$json = const {
  '1': 'QuerySwapExactAmountInRequest',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'poolId', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'tokenIn', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'tokenIn'},
    const {'1': 'routes', '3': 4, '4': 3, '5': 11, '6': '.osmosis.gamm.v1beta1.SwapAmountInRoute', '8': const {}, '10': 'routes'},
  ],
};

const QuerySwapExactAmountInResponse$json = const {
  '1': 'QuerySwapExactAmountInResponse',
  '2': const [
    const {'1': 'tokenOutAmount', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'tokenOutAmount'},
  ],
};

const QuerySwapExactAmountOutRequest$json = const {
  '1': 'QuerySwapExactAmountOutRequest',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'poolId', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'routes', '3': 3, '4': 3, '5': 11, '6': '.osmosis.gamm.v1beta1.SwapAmountOutRoute', '8': const {}, '10': 'routes'},
    const {'1': 'tokenOut', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'tokenOut'},
  ],
};

const QuerySwapExactAmountOutResponse$json = const {
  '1': 'QuerySwapExactAmountOutResponse',
  '2': const [
    const {'1': 'tokenInAmount', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'tokenInAmount'},
  ],
};

const QueryTotalLiquidityRequest$json = const {
  '1': 'QueryTotalLiquidityRequest',
};

const QueryTotalLiquidityResponse$json = const {
  '1': 'QueryTotalLiquidityResponse',
  '2': const [
    const {'1': 'liquidity', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'liquidity'},
  ],
};

