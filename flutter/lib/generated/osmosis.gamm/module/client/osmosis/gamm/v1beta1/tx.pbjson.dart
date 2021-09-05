///
//  Generated code. Do not modify.
//  source: osmosis/gamm/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const MsgCreatePool$json = const {
  '1': 'MsgCreatePool',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'poolParams', '3': 2, '4': 1, '5': 11, '6': '.osmosis.gamm.v1beta1.PoolParams', '8': const {}, '10': 'poolParams'},
    const {'1': 'poolAssets', '3': 3, '4': 3, '5': 11, '6': '.osmosis.gamm.v1beta1.PoolAsset', '8': const {}, '10': 'poolAssets'},
    const {'1': 'future_pool_governor', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'futurePoolGovernor'},
  ],
};

const MsgCreatePoolResponse$json = const {
  '1': 'MsgCreatePoolResponse',
};

const MsgJoinPool$json = const {
  '1': 'MsgJoinPool',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'poolId', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'shareOutAmount', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'shareOutAmount'},
    const {'1': 'tokenInMaxs', '3': 4, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'tokenInMaxs'},
  ],
};

const MsgJoinPoolResponse$json = const {
  '1': 'MsgJoinPoolResponse',
};

const MsgExitPool$json = const {
  '1': 'MsgExitPool',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'poolId', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'shareInAmount', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'shareInAmount'},
    const {'1': 'tokenOutMins', '3': 4, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'tokenOutMins'},
  ],
};

const MsgExitPoolResponse$json = const {
  '1': 'MsgExitPoolResponse',
};

const SwapAmountInRoute$json = const {
  '1': 'SwapAmountInRoute',
  '2': const [
    const {'1': 'poolId', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'tokenOutDenom', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'tokenOutDenom'},
  ],
};

const MsgSwapExactAmountIn$json = const {
  '1': 'MsgSwapExactAmountIn',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'routes', '3': 2, '4': 3, '5': 11, '6': '.osmosis.gamm.v1beta1.SwapAmountInRoute', '8': const {}, '10': 'routes'},
    const {'1': 'tokenIn', '3': 3, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'tokenIn'},
    const {'1': 'tokenOutMinAmount', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'tokenOutMinAmount'},
  ],
};

const MsgSwapExactAmountInResponse$json = const {
  '1': 'MsgSwapExactAmountInResponse',
};

const SwapAmountOutRoute$json = const {
  '1': 'SwapAmountOutRoute',
  '2': const [
    const {'1': 'poolId', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'tokenInDenom', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'tokenInDenom'},
  ],
};

const MsgSwapExactAmountOut$json = const {
  '1': 'MsgSwapExactAmountOut',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'routes', '3': 2, '4': 3, '5': 11, '6': '.osmosis.gamm.v1beta1.SwapAmountOutRoute', '8': const {}, '10': 'routes'},
    const {'1': 'tokenInMaxAmount', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'tokenInMaxAmount'},
    const {'1': 'tokenOut', '3': 4, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'tokenOut'},
  ],
};

const MsgSwapExactAmountOutResponse$json = const {
  '1': 'MsgSwapExactAmountOutResponse',
};

const MsgJoinSwapExternAmountIn$json = const {
  '1': 'MsgJoinSwapExternAmountIn',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'poolId', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'tokenIn', '3': 3, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'tokenIn'},
    const {'1': 'shareOutMinAmount', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'shareOutMinAmount'},
  ],
};

const MsgJoinSwapExternAmountInResponse$json = const {
  '1': 'MsgJoinSwapExternAmountInResponse',
};

const MsgJoinSwapShareAmountOut$json = const {
  '1': 'MsgJoinSwapShareAmountOut',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'poolId', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'tokenInDenom', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'tokenInDenom'},
    const {'1': 'shareOutAmount', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'shareOutAmount'},
    const {'1': 'tokenInMaxAmount', '3': 5, '4': 1, '5': 9, '8': const {}, '10': 'tokenInMaxAmount'},
  ],
};

const MsgJoinSwapShareAmountOutResponse$json = const {
  '1': 'MsgJoinSwapShareAmountOutResponse',
};

const MsgExitSwapShareAmountIn$json = const {
  '1': 'MsgExitSwapShareAmountIn',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'poolId', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'tokenOutDenom', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'tokenOutDenom'},
    const {'1': 'shareInAmount', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'shareInAmount'},
    const {'1': 'tokenOutMinAmount', '3': 5, '4': 1, '5': 9, '8': const {}, '10': 'tokenOutMinAmount'},
  ],
};

const MsgExitSwapShareAmountInResponse$json = const {
  '1': 'MsgExitSwapShareAmountInResponse',
};

const MsgExitSwapExternAmountOut$json = const {
  '1': 'MsgExitSwapExternAmountOut',
  '2': const [
    const {'1': 'sender', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'sender'},
    const {'1': 'poolId', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'poolId'},
    const {'1': 'tokenOut', '3': 3, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'tokenOut'},
    const {'1': 'shareInMaxAmount', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'shareInMaxAmount'},
  ],
};

const MsgExitSwapExternAmountOutResponse$json = const {
  '1': 'MsgExitSwapExternAmountOutResponse',
};

