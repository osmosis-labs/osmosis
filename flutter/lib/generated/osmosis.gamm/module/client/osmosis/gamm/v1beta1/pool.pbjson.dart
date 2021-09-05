///
//  Generated code. Do not modify.
//  source: osmosis/gamm/v1beta1/pool.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const PoolAsset$json = const {
  '1': 'PoolAsset',
  '2': const [
    const {'1': 'token', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'token'},
    const {'1': 'weight', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'weight'},
  ],
};

const SmoothWeightChangeParams$json = const {
  '1': 'SmoothWeightChangeParams',
  '2': const [
    const {'1': 'start_time', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'startTime'},
    const {'1': 'duration', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'duration'},
    const {'1': 'initialPoolWeights', '3': 3, '4': 3, '5': 11, '6': '.osmosis.gamm.v1beta1.PoolAsset', '8': const {}, '10': 'initialPoolWeights'},
    const {'1': 'targetPoolWeights', '3': 4, '4': 3, '5': 11, '6': '.osmosis.gamm.v1beta1.PoolAsset', '8': const {}, '10': 'targetPoolWeights'},
  ],
};

const PoolParams$json = const {
  '1': 'PoolParams',
  '2': const [
    const {'1': 'swapFee', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'swapFee'},
    const {'1': 'exitFee', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'exitFee'},
    const {'1': 'smoothWeightChangeParams', '3': 3, '4': 1, '5': 11, '6': '.osmosis.gamm.v1beta1.SmoothWeightChangeParams', '8': const {}, '10': 'smoothWeightChangeParams'},
  ],
};

const Pool$json = const {
  '1': 'Pool',
  '2': const [
    const {'1': 'address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'address'},
    const {'1': 'id', '3': 2, '4': 1, '5': 4, '10': 'id'},
    const {'1': 'poolParams', '3': 3, '4': 1, '5': 11, '6': '.osmosis.gamm.v1beta1.PoolParams', '8': const {}, '10': 'poolParams'},
    const {'1': 'future_pool_governor', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'futurePoolGovernor'},
    const {'1': 'totalShares', '3': 5, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'totalShares'},
    const {'1': 'poolAssets', '3': 6, '4': 3, '5': 11, '6': '.osmosis.gamm.v1beta1.PoolAsset', '8': const {}, '10': 'poolAssets'},
    const {'1': 'totalWeight', '3': 7, '4': 1, '5': 9, '8': const {}, '10': 'totalWeight'},
  ],
  '7': const {},
};

