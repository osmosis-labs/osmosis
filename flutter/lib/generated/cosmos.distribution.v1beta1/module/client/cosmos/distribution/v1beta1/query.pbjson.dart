///
//  Generated code. Do not modify.
//  source: cosmos/distribution/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const QueryParamsRequest$json = const {
  '1': 'QueryParamsRequest',
};

const QueryParamsResponse$json = const {
  '1': 'QueryParamsResponse',
  '2': const [
    const {'1': 'params', '3': 1, '4': 1, '5': 11, '6': '.cosmos.distribution.v1beta1.Params', '8': const {}, '10': 'params'},
  ],
};

const QueryValidatorOutstandingRewardsRequest$json = const {
  '1': 'QueryValidatorOutstandingRewardsRequest',
  '2': const [
    const {'1': 'validator_address', '3': 1, '4': 1, '5': 9, '10': 'validatorAddress'},
  ],
};

const QueryValidatorOutstandingRewardsResponse$json = const {
  '1': 'QueryValidatorOutstandingRewardsResponse',
  '2': const [
    const {'1': 'rewards', '3': 1, '4': 1, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorOutstandingRewards', '8': const {}, '10': 'rewards'},
  ],
};

const QueryValidatorCommissionRequest$json = const {
  '1': 'QueryValidatorCommissionRequest',
  '2': const [
    const {'1': 'validator_address', '3': 1, '4': 1, '5': 9, '10': 'validatorAddress'},
  ],
};

const QueryValidatorCommissionResponse$json = const {
  '1': 'QueryValidatorCommissionResponse',
  '2': const [
    const {'1': 'commission', '3': 1, '4': 1, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorAccumulatedCommission', '8': const {}, '10': 'commission'},
  ],
};

const QueryValidatorSlashesRequest$json = const {
  '1': 'QueryValidatorSlashesRequest',
  '2': const [
    const {'1': 'validator_address', '3': 1, '4': 1, '5': 9, '10': 'validatorAddress'},
    const {'1': 'starting_height', '3': 2, '4': 1, '5': 4, '10': 'startingHeight'},
    const {'1': 'ending_height', '3': 3, '4': 1, '5': 4, '10': 'endingHeight'},
    const {'1': 'pagination', '3': 4, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
  '7': const {},
};

const QueryValidatorSlashesResponse$json = const {
  '1': 'QueryValidatorSlashesResponse',
  '2': const [
    const {'1': 'slashes', '3': 1, '4': 3, '5': 11, '6': '.cosmos.distribution.v1beta1.ValidatorSlashEvent', '8': const {}, '10': 'slashes'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryDelegationRewardsRequest$json = const {
  '1': 'QueryDelegationRewardsRequest',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddress'},
    const {'1': 'validator_address', '3': 2, '4': 1, '5': 9, '10': 'validatorAddress'},
  ],
  '7': const {},
};

const QueryDelegationRewardsResponse$json = const {
  '1': 'QueryDelegationRewardsResponse',
  '2': const [
    const {'1': 'rewards', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.DecCoin', '8': const {}, '10': 'rewards'},
  ],
};

const QueryDelegationTotalRewardsRequest$json = const {
  '1': 'QueryDelegationTotalRewardsRequest',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddress'},
  ],
  '7': const {},
};

const QueryDelegationTotalRewardsResponse$json = const {
  '1': 'QueryDelegationTotalRewardsResponse',
  '2': const [
    const {'1': 'rewards', '3': 1, '4': 3, '5': 11, '6': '.cosmos.distribution.v1beta1.DelegationDelegatorReward', '8': const {}, '10': 'rewards'},
    const {'1': 'total', '3': 2, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.DecCoin', '8': const {}, '10': 'total'},
  ],
};

const QueryDelegatorValidatorsRequest$json = const {
  '1': 'QueryDelegatorValidatorsRequest',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddress'},
  ],
  '7': const {},
};

const QueryDelegatorValidatorsResponse$json = const {
  '1': 'QueryDelegatorValidatorsResponse',
  '2': const [
    const {'1': 'validators', '3': 1, '4': 3, '5': 9, '10': 'validators'},
  ],
  '7': const {},
};

const QueryDelegatorWithdrawAddressRequest$json = const {
  '1': 'QueryDelegatorWithdrawAddressRequest',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddress'},
  ],
  '7': const {},
};

const QueryDelegatorWithdrawAddressResponse$json = const {
  '1': 'QueryDelegatorWithdrawAddressResponse',
  '2': const [
    const {'1': 'withdraw_address', '3': 1, '4': 1, '5': 9, '10': 'withdrawAddress'},
  ],
  '7': const {},
};

const QueryCommunityPoolRequest$json = const {
  '1': 'QueryCommunityPoolRequest',
};

const QueryCommunityPoolResponse$json = const {
  '1': 'QueryCommunityPoolResponse',
  '2': const [
    const {'1': 'pool', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.DecCoin', '8': const {}, '10': 'pool'},
  ],
};

