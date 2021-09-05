///
//  Generated code. Do not modify.
//  source: osmosis/lockup/lock.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const LockQueryType$json = const {
  '1': 'LockQueryType',
  '2': const [
    const {'1': 'ByDuration', '2': 0},
    const {'1': 'ByTime', '2': 1},
  ],
  '3': const {},
};

const PeriodLock$json = const {
  '1': 'PeriodLock',
  '2': const [
    const {'1': 'ID', '3': 1, '4': 1, '5': 4, '10': 'ID'},
    const {'1': 'owner', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'duration', '3': 3, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'duration'},
    const {'1': 'end_time', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'endTime'},
    const {'1': 'coins', '3': 5, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
};

const QueryCondition$json = const {
  '1': 'QueryCondition',
  '2': const [
    const {'1': 'lock_query_type', '3': 1, '4': 1, '5': 14, '6': '.osmosis.lockup.LockQueryType', '10': 'lockQueryType'},
    const {'1': 'denom', '3': 2, '4': 1, '5': 9, '10': 'denom'},
    const {'1': 'duration', '3': 3, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'duration'},
    const {'1': 'timestamp', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'timestamp'},
  ],
};

