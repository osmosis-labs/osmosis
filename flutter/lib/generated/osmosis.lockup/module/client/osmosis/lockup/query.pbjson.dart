///
//  Generated code. Do not modify.
//  source: osmosis/lockup/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const ModuleBalanceRequest$json = const {
  '1': 'ModuleBalanceRequest',
};

const ModuleBalanceResponse$json = const {
  '1': 'ModuleBalanceResponse',
  '2': const [
    const {'1': 'coins', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
};

const ModuleLockedAmountRequest$json = const {
  '1': 'ModuleLockedAmountRequest',
};

const ModuleLockedAmountResponse$json = const {
  '1': 'ModuleLockedAmountResponse',
  '2': const [
    const {'1': 'coins', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
};

const AccountUnlockableCoinsRequest$json = const {
  '1': 'AccountUnlockableCoinsRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
  ],
};

const AccountUnlockableCoinsResponse$json = const {
  '1': 'AccountUnlockableCoinsResponse',
  '2': const [
    const {'1': 'coins', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
};

const AccountUnlockingCoinsRequest$json = const {
  '1': 'AccountUnlockingCoinsRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
  ],
};

const AccountUnlockingCoinsResponse$json = const {
  '1': 'AccountUnlockingCoinsResponse',
  '2': const [
    const {'1': 'coins', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
};

const AccountLockedCoinsRequest$json = const {
  '1': 'AccountLockedCoinsRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
  ],
};

const AccountLockedCoinsResponse$json = const {
  '1': 'AccountLockedCoinsResponse',
  '2': const [
    const {'1': 'coins', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
};

const AccountLockedPastTimeRequest$json = const {
  '1': 'AccountLockedPastTimeRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'timestamp', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'timestamp'},
  ],
};

const AccountLockedPastTimeResponse$json = const {
  '1': 'AccountLockedPastTimeResponse',
  '2': const [
    const {'1': 'locks', '3': 1, '4': 3, '5': 11, '6': '.osmosis.lockup.PeriodLock', '8': const {}, '10': 'locks'},
  ],
};

const AccountLockedPastTimeNotUnlockingOnlyRequest$json = const {
  '1': 'AccountLockedPastTimeNotUnlockingOnlyRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'timestamp', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'timestamp'},
  ],
};

const AccountLockedPastTimeNotUnlockingOnlyResponse$json = const {
  '1': 'AccountLockedPastTimeNotUnlockingOnlyResponse',
  '2': const [
    const {'1': 'locks', '3': 1, '4': 3, '5': 11, '6': '.osmosis.lockup.PeriodLock', '8': const {}, '10': 'locks'},
  ],
};

const AccountUnlockedBeforeTimeRequest$json = const {
  '1': 'AccountUnlockedBeforeTimeRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'timestamp', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'timestamp'},
  ],
};

const AccountUnlockedBeforeTimeResponse$json = const {
  '1': 'AccountUnlockedBeforeTimeResponse',
  '2': const [
    const {'1': 'locks', '3': 1, '4': 3, '5': 11, '6': '.osmosis.lockup.PeriodLock', '8': const {}, '10': 'locks'},
  ],
};

const AccountLockedPastTimeDenomRequest$json = const {
  '1': 'AccountLockedPastTimeDenomRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'timestamp', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'timestamp'},
    const {'1': 'denom', '3': 3, '4': 1, '5': 9, '10': 'denom'},
  ],
};

const AccountLockedPastTimeDenomResponse$json = const {
  '1': 'AccountLockedPastTimeDenomResponse',
  '2': const [
    const {'1': 'locks', '3': 1, '4': 3, '5': 11, '6': '.osmosis.lockup.PeriodLock', '8': const {}, '10': 'locks'},
  ],
};

const LockedDenomRequest$json = const {
  '1': 'LockedDenomRequest',
  '2': const [
    const {'1': 'denom', '3': 1, '4': 1, '5': 9, '10': 'denom'},
    const {'1': 'duration', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'duration'},
  ],
};

const LockedDenomResponse$json = const {
  '1': 'LockedDenomResponse',
  '2': const [
    const {'1': 'amount', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'amount'},
  ],
};

const LockedRequest$json = const {
  '1': 'LockedRequest',
  '2': const [
    const {'1': 'lock_id', '3': 1, '4': 1, '5': 4, '10': 'lockId'},
  ],
};

const LockedResponse$json = const {
  '1': 'LockedResponse',
  '2': const [
    const {'1': 'lock', '3': 1, '4': 1, '5': 11, '6': '.osmosis.lockup.PeriodLock', '10': 'lock'},
  ],
};

const AccountLockedLongerDurationRequest$json = const {
  '1': 'AccountLockedLongerDurationRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'duration', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'duration'},
  ],
};

const AccountLockedLongerDurationResponse$json = const {
  '1': 'AccountLockedLongerDurationResponse',
  '2': const [
    const {'1': 'locks', '3': 1, '4': 3, '5': 11, '6': '.osmosis.lockup.PeriodLock', '8': const {}, '10': 'locks'},
  ],
};

const AccountLockedLongerDurationNotUnlockingOnlyRequest$json = const {
  '1': 'AccountLockedLongerDurationNotUnlockingOnlyRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'duration', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'duration'},
  ],
};

const AccountLockedLongerDurationNotUnlockingOnlyResponse$json = const {
  '1': 'AccountLockedLongerDurationNotUnlockingOnlyResponse',
  '2': const [
    const {'1': 'locks', '3': 1, '4': 3, '5': 11, '6': '.osmosis.lockup.PeriodLock', '8': const {}, '10': 'locks'},
  ],
};

const AccountLockedLongerDurationDenomRequest$json = const {
  '1': 'AccountLockedLongerDurationDenomRequest',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'duration', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'duration'},
    const {'1': 'denom', '3': 3, '4': 1, '5': 9, '10': 'denom'},
  ],
};

const AccountLockedLongerDurationDenomResponse$json = const {
  '1': 'AccountLockedLongerDurationDenomResponse',
  '2': const [
    const {'1': 'locks', '3': 1, '4': 3, '5': 11, '6': '.osmosis.lockup.PeriodLock', '8': const {}, '10': 'locks'},
  ],
};

