///
//  Generated code. Do not modify.
//  source: cosmos/vesting/v1beta1/vesting.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const BaseVestingAccount$json = const {
  '1': 'BaseVestingAccount',
  '2': const [
    const {'1': 'base_account', '3': 1, '4': 1, '5': 11, '6': '.cosmos.auth.v1beta1.BaseAccount', '8': const {}, '10': 'baseAccount'},
    const {'1': 'original_vesting', '3': 2, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'originalVesting'},
    const {'1': 'delegated_free', '3': 3, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'delegatedFree'},
    const {'1': 'delegated_vesting', '3': 4, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'delegatedVesting'},
    const {'1': 'end_time', '3': 5, '4': 1, '5': 3, '8': const {}, '10': 'endTime'},
  ],
  '7': const {},
};

const ContinuousVestingAccount$json = const {
  '1': 'ContinuousVestingAccount',
  '2': const [
    const {'1': 'base_vesting_account', '3': 1, '4': 1, '5': 11, '6': '.cosmos.vesting.v1beta1.BaseVestingAccount', '8': const {}, '10': 'baseVestingAccount'},
    const {'1': 'start_time', '3': 2, '4': 1, '5': 3, '8': const {}, '10': 'startTime'},
  ],
  '7': const {},
};

const DelayedVestingAccount$json = const {
  '1': 'DelayedVestingAccount',
  '2': const [
    const {'1': 'base_vesting_account', '3': 1, '4': 1, '5': 11, '6': '.cosmos.vesting.v1beta1.BaseVestingAccount', '8': const {}, '10': 'baseVestingAccount'},
  ],
  '7': const {},
};

const Period$json = const {
  '1': 'Period',
  '2': const [
    const {'1': 'length', '3': 1, '4': 1, '5': 3, '10': 'length'},
    const {'1': 'amount', '3': 2, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'amount'},
  ],
  '7': const {},
};

const PeriodicVestingAccount$json = const {
  '1': 'PeriodicVestingAccount',
  '2': const [
    const {'1': 'base_vesting_account', '3': 1, '4': 1, '5': 11, '6': '.cosmos.vesting.v1beta1.BaseVestingAccount', '8': const {}, '10': 'baseVestingAccount'},
    const {'1': 'start_time', '3': 2, '4': 1, '5': 3, '8': const {}, '10': 'startTime'},
    const {'1': 'vesting_periods', '3': 3, '4': 3, '5': 11, '6': '.cosmos.vesting.v1beta1.Period', '8': const {}, '10': 'vestingPeriods'},
  ],
  '7': const {},
};

