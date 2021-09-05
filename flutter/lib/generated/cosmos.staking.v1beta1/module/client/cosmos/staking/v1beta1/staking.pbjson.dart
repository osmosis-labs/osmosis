///
//  Generated code. Do not modify.
//  source: cosmos/staking/v1beta1/staking.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const BondStatus$json = const {
  '1': 'BondStatus',
  '2': const [
    const {'1': 'BOND_STATUS_UNSPECIFIED', '2': 0, '3': const {}},
    const {'1': 'BOND_STATUS_UNBONDED', '2': 1, '3': const {}},
    const {'1': 'BOND_STATUS_UNBONDING', '2': 2, '3': const {}},
    const {'1': 'BOND_STATUS_BONDED', '2': 3, '3': const {}},
  ],
  '3': const {},
};

const HistoricalInfo$json = const {
  '1': 'HistoricalInfo',
  '2': const [
    const {'1': 'header', '3': 1, '4': 1, '5': 11, '6': '.tendermint.types.Header', '8': const {}, '10': 'header'},
    const {'1': 'valset', '3': 2, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.Validator', '8': const {}, '10': 'valset'},
  ],
};

const CommissionRates$json = const {
  '1': 'CommissionRates',
  '2': const [
    const {'1': 'rate', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'rate'},
    const {'1': 'max_rate', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'maxRate'},
    const {'1': 'max_change_rate', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'maxChangeRate'},
  ],
  '7': const {},
};

const Commission$json = const {
  '1': 'Commission',
  '2': const [
    const {'1': 'commission_rates', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.CommissionRates', '8': const {}, '10': 'commissionRates'},
    const {'1': 'update_time', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'updateTime'},
  ],
  '7': const {},
};

const Description$json = const {
  '1': 'Description',
  '2': const [
    const {'1': 'moniker', '3': 1, '4': 1, '5': 9, '10': 'moniker'},
    const {'1': 'identity', '3': 2, '4': 1, '5': 9, '10': 'identity'},
    const {'1': 'website', '3': 3, '4': 1, '5': 9, '10': 'website'},
    const {'1': 'security_contact', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'securityContact'},
    const {'1': 'details', '3': 5, '4': 1, '5': 9, '10': 'details'},
  ],
  '7': const {},
};

const Validator$json = const {
  '1': 'Validator',
  '2': const [
    const {'1': 'operator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'operatorAddress'},
    const {'1': 'consensus_pubkey', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'consensusPubkey'},
    const {'1': 'jailed', '3': 3, '4': 1, '5': 8, '10': 'jailed'},
    const {'1': 'status', '3': 4, '4': 1, '5': 14, '6': '.cosmos.staking.v1beta1.BondStatus', '10': 'status'},
    const {'1': 'tokens', '3': 5, '4': 1, '5': 9, '8': const {}, '10': 'tokens'},
    const {'1': 'delegator_shares', '3': 6, '4': 1, '5': 9, '8': const {}, '10': 'delegatorShares'},
    const {'1': 'description', '3': 7, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.Description', '8': const {}, '10': 'description'},
    const {'1': 'unbonding_height', '3': 8, '4': 1, '5': 3, '8': const {}, '10': 'unbondingHeight'},
    const {'1': 'unbonding_time', '3': 9, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'unbondingTime'},
    const {'1': 'commission', '3': 10, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.Commission', '8': const {}, '10': 'commission'},
    const {'1': 'min_self_delegation', '3': 11, '4': 1, '5': 9, '8': const {}, '10': 'minSelfDelegation'},
  ],
  '7': const {},
};

const ValAddresses$json = const {
  '1': 'ValAddresses',
  '2': const [
    const {'1': 'addresses', '3': 1, '4': 3, '5': 9, '10': 'addresses'},
  ],
  '7': const {},
};

const DVPair$json = const {
  '1': 'DVPair',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
  ],
  '7': const {},
};

const DVPairs$json = const {
  '1': 'DVPairs',
  '2': const [
    const {'1': 'pairs', '3': 1, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.DVPair', '8': const {}, '10': 'pairs'},
  ],
};

const DVVTriplet$json = const {
  '1': 'DVVTriplet',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_src_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorSrcAddress'},
    const {'1': 'validator_dst_address', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'validatorDstAddress'},
  ],
  '7': const {},
};

const DVVTriplets$json = const {
  '1': 'DVVTriplets',
  '2': const [
    const {'1': 'triplets', '3': 1, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.DVVTriplet', '8': const {}, '10': 'triplets'},
  ],
};

const Delegation$json = const {
  '1': 'Delegation',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'shares', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'shares'},
  ],
  '7': const {},
};

const UnbondingDelegation$json = const {
  '1': 'UnbondingDelegation',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'entries', '3': 3, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.UnbondingDelegationEntry', '8': const {}, '10': 'entries'},
  ],
  '7': const {},
};

const UnbondingDelegationEntry$json = const {
  '1': 'UnbondingDelegationEntry',
  '2': const [
    const {'1': 'creation_height', '3': 1, '4': 1, '5': 3, '8': const {}, '10': 'creationHeight'},
    const {'1': 'completion_time', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'completionTime'},
    const {'1': 'initial_balance', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'initialBalance'},
    const {'1': 'balance', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'balance'},
  ],
  '7': const {},
};

const RedelegationEntry$json = const {
  '1': 'RedelegationEntry',
  '2': const [
    const {'1': 'creation_height', '3': 1, '4': 1, '5': 3, '8': const {}, '10': 'creationHeight'},
    const {'1': 'completion_time', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'completionTime'},
    const {'1': 'initial_balance', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'initialBalance'},
    const {'1': 'shares_dst', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'sharesDst'},
  ],
  '7': const {},
};

const Redelegation$json = const {
  '1': 'Redelegation',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_src_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorSrcAddress'},
    const {'1': 'validator_dst_address', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'validatorDstAddress'},
    const {'1': 'entries', '3': 4, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.RedelegationEntry', '8': const {}, '10': 'entries'},
  ],
  '7': const {},
};

const Params$json = const {
  '1': 'Params',
  '2': const [
    const {'1': 'unbonding_time', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'unbondingTime'},
    const {'1': 'max_validators', '3': 2, '4': 1, '5': 13, '8': const {}, '10': 'maxValidators'},
    const {'1': 'max_entries', '3': 3, '4': 1, '5': 13, '8': const {}, '10': 'maxEntries'},
    const {'1': 'historical_entries', '3': 4, '4': 1, '5': 13, '8': const {}, '10': 'historicalEntries'},
    const {'1': 'bond_denom', '3': 5, '4': 1, '5': 9, '8': const {}, '10': 'bondDenom'},
    const {'1': 'min_commission_rate', '3': 6, '4': 1, '5': 9, '8': const {}, '10': 'minCommissionRate'},
  ],
  '7': const {},
};

const DelegationResponse$json = const {
  '1': 'DelegationResponse',
  '2': const [
    const {'1': 'delegation', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.Delegation', '8': const {}, '10': 'delegation'},
    const {'1': 'balance', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'balance'},
  ],
  '7': const {},
};

const RedelegationEntryResponse$json = const {
  '1': 'RedelegationEntryResponse',
  '2': const [
    const {'1': 'redelegation_entry', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.RedelegationEntry', '8': const {}, '10': 'redelegationEntry'},
    const {'1': 'balance', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'balance'},
  ],
  '7': const {},
};

const RedelegationResponse$json = const {
  '1': 'RedelegationResponse',
  '2': const [
    const {'1': 'redelegation', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.Redelegation', '8': const {}, '10': 'redelegation'},
    const {'1': 'entries', '3': 2, '4': 3, '5': 11, '6': '.cosmos.staking.v1beta1.RedelegationEntryResponse', '8': const {}, '10': 'entries'},
  ],
  '7': const {},
};

const Pool$json = const {
  '1': 'Pool',
  '2': const [
    const {'1': 'not_bonded_tokens', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'notBondedTokens'},
    const {'1': 'bonded_tokens', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'bondedTokens'},
  ],
  '7': const {},
};

