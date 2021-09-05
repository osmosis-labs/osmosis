///
//  Generated code. Do not modify.
//  source: osmosis/lockup/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const MsgLockTokens$json = const {
  '1': 'MsgLockTokens',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'duration', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'duration'},
    const {'1': 'coins', '3': 3, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
};

const MsgLockTokensResponse$json = const {
  '1': 'MsgLockTokensResponse',
  '2': const [
    const {'1': 'ID', '3': 1, '4': 1, '5': 4, '10': 'ID'},
  ],
};

const MsgBeginUnlockingAll$json = const {
  '1': 'MsgBeginUnlockingAll',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
  ],
};

const MsgBeginUnlockingAllResponse$json = const {
  '1': 'MsgBeginUnlockingAllResponse',
  '2': const [
    const {'1': 'unlocks', '3': 1, '4': 3, '5': 11, '6': '.osmosis.lockup.PeriodLock', '10': 'unlocks'},
  ],
};

const MsgBeginUnlocking$json = const {
  '1': 'MsgBeginUnlocking',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'ID', '3': 2, '4': 1, '5': 4, '10': 'ID'},
  ],
};

const MsgBeginUnlockingResponse$json = const {
  '1': 'MsgBeginUnlockingResponse',
  '2': const [
    const {'1': 'success', '3': 1, '4': 1, '5': 8, '10': 'success'},
  ],
};

