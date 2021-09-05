///
//  Generated code. Do not modify.
//  source: cosmos/staking/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const QueryValidatorsRequest$json = const {
  '1': 'QueryValidatorsRequest',
  '2': const [
    const {'1': 'status', '3': 1, '4': 1, '5': 9, '10': 'status'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryValidatorsResponse$json = const {
  '1': 'QueryValidatorsResponse',
  '2': const [
    const {'1': 'validators', '3': 1, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.Validator', '8': const {}, '10': 'validators'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryValidatorRequest$json = const {
  '1': 'QueryValidatorRequest',
  '2': const [
    const {'1': 'validator_addr', '3': 1, '4': 1, '5': 9, '10': 'validatorAddr'},
  ],
};

const QueryValidatorResponse$json = const {
  '1': 'QueryValidatorResponse',
  '2': const [
    const {'1': 'validator', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.Validator', '8': const {}, '10': 'validator'},
  ],
};

const QueryValidatorDelegationsRequest$json = const {
  '1': 'QueryValidatorDelegationsRequest',
  '2': const [
    const {'1': 'validator_addr', '3': 1, '4': 1, '5': 9, '10': 'validatorAddr'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryValidatorDelegationsResponse$json = const {
  '1': 'QueryValidatorDelegationsResponse',
  '2': const [
    const {'1': 'delegation_responses', '3': 1, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.DelegationResponse', '8': const {}, '10': 'delegationResponses'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryValidatorUnbondingDelegationsRequest$json = const {
  '1': 'QueryValidatorUnbondingDelegationsRequest',
  '2': const [
    const {'1': 'validator_addr', '3': 1, '4': 1, '5': 9, '10': 'validatorAddr'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryValidatorUnbondingDelegationsResponse$json = const {
  '1': 'QueryValidatorUnbondingDelegationsResponse',
  '2': const [
    const {'1': 'unbonding_responses', '3': 1, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.UnbondingDelegation', '8': const {}, '10': 'unbondingResponses'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryDelegationRequest$json = const {
  '1': 'QueryDelegationRequest',
  '2': const [
    const {'1': 'delegator_addr', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddr'},
    const {'1': 'validator_addr', '3': 2, '4': 1, '5': 9, '10': 'validatorAddr'},
  ],
  '7': const {},
};

const QueryDelegationResponse$json = const {
  '1': 'QueryDelegationResponse',
  '2': const [
    const {'1': 'delegation_response', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.DelegationResponse', '10': 'delegationResponse'},
  ],
};

const QueryUnbondingDelegationRequest$json = const {
  '1': 'QueryUnbondingDelegationRequest',
  '2': const [
    const {'1': 'delegator_addr', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddr'},
    const {'1': 'validator_addr', '3': 2, '4': 1, '5': 9, '10': 'validatorAddr'},
  ],
  '7': const {},
};

const QueryUnbondingDelegationResponse$json = const {
  '1': 'QueryUnbondingDelegationResponse',
  '2': const [
    const {'1': 'unbond', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.UnbondingDelegation', '8': const {}, '10': 'unbond'},
  ],
};

const QueryDelegatorDelegationsRequest$json = const {
  '1': 'QueryDelegatorDelegationsRequest',
  '2': const [
    const {'1': 'delegator_addr', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddr'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
  '7': const {},
};

const QueryDelegatorDelegationsResponse$json = const {
  '1': 'QueryDelegatorDelegationsResponse',
  '2': const [
    const {'1': 'delegation_responses', '3': 1, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.DelegationResponse', '8': const {}, '10': 'delegationResponses'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryDelegatorUnbondingDelegationsRequest$json = const {
  '1': 'QueryDelegatorUnbondingDelegationsRequest',
  '2': const [
    const {'1': 'delegator_addr', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddr'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
  '7': const {},
};

const QueryDelegatorUnbondingDelegationsResponse$json = const {
  '1': 'QueryDelegatorUnbondingDelegationsResponse',
  '2': const [
    const {'1': 'unbonding_responses', '3': 1, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.UnbondingDelegation', '8': const {}, '10': 'unbondingResponses'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryRedelegationsRequest$json = const {
  '1': 'QueryRedelegationsRequest',
  '2': const [
    const {'1': 'delegator_addr', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddr'},
    const {'1': 'src_validator_addr', '3': 2, '4': 1, '5': 9, '10': 'srcValidatorAddr'},
    const {'1': 'dst_validator_addr', '3': 3, '4': 1, '5': 9, '10': 'dstValidatorAddr'},
    const {'1': 'pagination', '3': 4, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
  '7': const {},
};

const QueryRedelegationsResponse$json = const {
  '1': 'QueryRedelegationsResponse',
  '2': const [
    const {'1': 'redelegation_responses', '3': 1, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.RedelegationResponse', '8': const {}, '10': 'redelegationResponses'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryDelegatorValidatorsRequest$json = const {
  '1': 'QueryDelegatorValidatorsRequest',
  '2': const [
    const {'1': 'delegator_addr', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddr'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
  '7': const {},
};

const QueryDelegatorValidatorsResponse$json = const {
  '1': 'QueryDelegatorValidatorsResponse',
  '2': const [
    const {'1': 'validators', '3': 1, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.Validator', '8': const {}, '10': 'validators'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryDelegatorValidatorRequest$json = const {
  '1': 'QueryDelegatorValidatorRequest',
  '2': const [
    const {'1': 'delegator_addr', '3': 1, '4': 1, '5': 9, '10': 'delegatorAddr'},
    const {'1': 'validator_addr', '3': 2, '4': 1, '5': 9, '10': 'validatorAddr'},
  ],
  '7': const {},
};

const QueryDelegatorValidatorResponse$json = const {
  '1': 'QueryDelegatorValidatorResponse',
  '2': const [
    const {'1': 'validator', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.Validator', '8': const {}, '10': 'validator'},
  ],
};

const QueryHistoricalInfoRequest$json = const {
  '1': 'QueryHistoricalInfoRequest',
  '2': const [
    const {'1': 'height', '3': 1, '4': 1, '5': 3, '10': 'height'},
  ],
};

const QueryHistoricalInfoResponse$json = const {
  '1': 'QueryHistoricalInfoResponse',
  '2': const [
    const {'1': 'hist', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.HistoricalInfo', '10': 'hist'},
  ],
};

const QueryPoolRequest$json = const {
  '1': 'QueryPoolRequest',
};

const QueryPoolResponse$json = const {
  '1': 'QueryPoolResponse',
  '2': const [
    const {'1': 'pool', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.Pool', '8': const {}, '10': 'pool'},
  ],
};

const QueryParamsRequest$json = const {
  '1': 'QueryParamsRequest',
};

const QueryParamsResponse$json = const {
  '1': 'QueryParamsResponse',
  '2': const [
    const {'1': 'params', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.Params', '8': const {}, '10': 'params'},
  ],
};

