///
//  Generated code. Do not modify.
//  source: cosmos/distribution/v1beta1/genesis.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const DelegatorWithdrawInfo$json = const {
  '1': 'DelegatorWithdrawInfo',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'withdraw_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'withdrawAddress'},
  ],
  '7': const {},
};

const ValidatorOutstandingRewardsRecord$json = const {
  '1': 'ValidatorOutstandingRewardsRecord',
  '2': const [
    const {'1': 'validator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'outstanding_rewards', '3': 2, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.DecCoin', '8': const {}, '10': 'outstandingRewards'},
  ],
  '7': const {},
};

const ValidatorAccumulatedCommissionRecord$json = const {
  '1': 'ValidatorAccumulatedCommissionRecord',
  '2': const [
    const {'1': 'validator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'accumulated', '3': 2, '4': 1, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorAccumulatedCommission', '8': const {}, '10': 'accumulated'},
  ],
  '7': const {},
};

const ValidatorHistoricalRewardsRecord$json = const {
  '1': 'ValidatorHistoricalRewardsRecord',
  '2': const [
    const {'1': 'validator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'period', '3': 2, '4': 1, '5': 4, '10': 'period'},
    const {'1': 'rewards', '3': 3, '4': 1, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorHistoricalRewards', '8': const {}, '10': 'rewards'},
  ],
  '7': const {},
};

const ValidatorCurrentRewardsRecord$json = const {
  '1': 'ValidatorCurrentRewardsRecord',
  '2': const [
    const {'1': 'validator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'rewards', '3': 2, '4': 1, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorCurrentRewards', '8': const {}, '10': 'rewards'},
  ],
  '7': const {},
};

const DelegatorStartingInfoRecord$json = const {
  '1': 'DelegatorStartingInfoRecord',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'starting_info', '3': 3, '4': 1, '5': 11, '6': '.cosmos.distribution.v1beta1.DelegatorStartingInfo', '8': const {}, '10': 'startingInfo'},
  ],
  '7': const {},
};

const ValidatorSlashEventRecord$json = const {
  '1': 'ValidatorSlashEventRecord',
  '2': const [
    const {'1': 'validator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'height', '3': 2, '4': 1, '5': 4, '10': 'height'},
    const {'1': 'period', '3': 3, '4': 1, '5': 4, '10': 'period'},
    const {'1': 'validator_slash_event', '3': 4, '4': 1, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorSlashEvent', '8': const {}, '10': 'validatorSlashEvent'},
  ],
  '7': const {},
};

const GenesisState$json = const {
  '1': 'GenesisState',
  '2': const [
    const {'1': 'params', '3': 1, '4': 1, '5': 11, '6': '.cosmos.distribution.v1beta1.Params', '8': const {}, '10': 'params'},
    const {'1': 'fee_pool', '3': 2, '4': 1, '5': 11, '6': '.cosmos.distribution.v1beta1.FeePool', '8': const {}, '10': 'feePool'},
    const {'1': 'delegator_withdraw_infos', '3': 3, '4': 3, '5': 11, '6': '.cosmos.distribution.v1beta1.DelegatorWithdrawInfo', '8': const {}, '10': 'delegatorWithdrawInfos'},
    const {'1': 'previous_proposer', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'previousProposer'},
    const {'1': 'outstanding_rewards', '3': 5, '4': 3, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorOutstandingRewardsRecord', '8': const {}, '10': 'outstandingRewards'},
    const {'1': 'validator_accumulated_commissions', '3': 6, '4': 3, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorAccumulatedCommissionRecord', '8': const {}, '10': 'validatorAccumulatedCommissions'},
    const {'1': 'validator_historical_rewards', '3': 7, '4': 3, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorHistoricalRewardsRecord', '8': const {}, '10': 'validatorHistoricalRewards'},
    const {'1': 'validator_current_rewards', '3': 8, '4': 3, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorCurrentRewardsRecord', '8': const {}, '10': 'validatorCurrentRewards'},
    const {'1': 'delegator_starting_infos', '3': 9, '4': 3, '5': 11, '6': '.cosmos.distribution.v1beta1.DelegatorStartingInfoRecord', '8': const {}, '10': 'delegatorStartingInfos'},
    const {'1': 'validator_slash_events', '3': 10, '4': 3, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorSlashEventRecord', '8': const {}, '10': 'validatorSlashEvents'},
  ],
  '7': const {},
};

