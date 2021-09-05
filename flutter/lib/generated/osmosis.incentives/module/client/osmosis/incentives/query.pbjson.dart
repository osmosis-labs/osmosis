///
//  Generated code. Do not modify.
//  source: osmosis/incentives/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const ModuleToDistributeCoinsRequest$json = const {
  '1': 'ModuleToDistributeCoinsRequest',
};

const ModuleToDistributeCoinsResponse$json = const {
  '1': 'ModuleToDistributeCoinsResponse',
  '2': const [
    const {'1': 'coins', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
};

const ModuleDistributedCoinsRequest$json = const {
  '1': 'ModuleDistributedCoinsRequest',
};

const ModuleDistributedCoinsResponse$json = const {
  '1': 'ModuleDistributedCoinsResponse',
  '2': const [
    const {'1': 'coins', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
};

const GaugeByIDRequest$json = const {
  '1': 'GaugeByIDRequest',
  '2': const [
    const {'1': 'id', '3': 1, '4': 1, '5': 4, '10': 'id'},
  ],
};

const GaugeByIDResponse$json = const {
  '1': 'GaugeByIDResponse',
  '2': const [
    const {'1': 'gauge', '3': 1, '4': 1, '5': 11, '6': '.osmosis.incentives.Gauge', '10': 'gauge'},
  ],
};

const GaugesRequest$json = const {
  '1': 'GaugesRequest',
  '2': const [
    const {'1': 'pagination', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const GaugesResponse$json = const {
  '1': 'GaugesResponse',
  '2': const [
    const {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.osmosis.incentives.Gauge', '8': const {}, '10': 'data'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const ActiveGaugesRequest$json = const {
  '1': 'ActiveGaugesRequest',
  '2': const [
    const {'1': 'pagination', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const ActiveGaugesResponse$json = const {
  '1': 'ActiveGaugesResponse',
  '2': const [
    const {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.osmosis.incentives.Gauge', '8': const {}, '10': 'data'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const UpcomingGaugesRequest$json = const {
  '1': 'UpcomingGaugesRequest',
  '2': const [
    const {'1': 'pagination', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const UpcomingGaugesResponse$json = const {
  '1': 'UpcomingGaugesResponse',
  '2': const [
    const {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.osmosis.incentives.Gauge', '8': const {}, '10': 'data'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const RewardsEstRequest$json = const {
  '1': 'RewardsEstRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'lock_ids', '3': 2, '4': 3, '5': 4, '10': 'lockIds'},
    const {'1': 'end_epoch', '3': 3, '4': 1, '5': 3, '10': 'endEpoch'},
  ],
};

const RewardsEstResponse$json = const {
  '1': 'RewardsEstResponse',
  '2': const [
    const {'1': 'coins', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
};

const QueryLockableDurationsRequest$json = const {
  '1': 'QueryLockableDurationsRequest',
};

const QueryLockableDurationsResponse$json = const {
  '1': 'QueryLockableDurationsResponse',
  '2': const [
    const {'1': 'lockable_durations', '3': 1, '4': 3, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'lockableDurations'},
  ],
};

