///
//  Generated code. Do not modify.
//  source: cosmos/staking/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const MsgCreateValidator$json = const {
  '1': 'MsgCreateValidator',
  '2': const [
    const {'1': 'description', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.Description', '8': const {}, '10': 'description'},
    const {'1': 'commission', '3': 2, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.CommissionRates', '8': const {}, '10': 'commission'},
    const {'1': 'min_self_delegation', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'minSelfDelegation'},
    const {'1': 'delegator_address', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_address', '3': 5, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'pubkey', '3': 6, '4': 1, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'pubkey'},
    const {'1': 'value', '3': 7, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'value'},
  ],
  '7': const {},
};

const MsgCreateValidatorResponse$json = const {
  '1': 'MsgCreateValidatorResponse',
};

const MsgEditValidator$json = const {
  '1': 'MsgEditValidator',
  '2': const [
    const {'1': 'description', '3': 1, '4': 1, '5': 11, '6': '.cosmos.staking.v1beta1.Description', '8': const {}, '10': 'description'},
    const {'1': 'validator_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'commission_rate', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'commissionRate'},
    const {'1': 'min_self_delegation', '3': 4, '4': 1, '5': 9, '8': const {}, '10': 'minSelfDelegation'},
  ],
  '7': const {},
};

const MsgEditValidatorResponse$json = const {
  '1': 'MsgEditValidatorResponse',
};

const MsgDelegate$json = const {
  '1': 'MsgDelegate',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'amount', '3': 3, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'amount'},
  ],
  '7': const {},
};

const MsgDelegateResponse$json = const {
  '1': 'MsgDelegateResponse',
};

const MsgBeginRedelegate$json = const {
  '1': 'MsgBeginRedelegate',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_src_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorSrcAddress'},
    const {'1': 'validator_dst_address', '3': 3, '4': 1, '5': 9, '8': const {}, '10': 'validatorDstAddress'},
    const {'1': 'amount', '3': 4, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'amount'},
  ],
  '7': const {},
};

const MsgBeginRedelegateResponse$json = const {
  '1': 'MsgBeginRedelegateResponse',
  '2': const [
    const {'1': 'completion_time', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'completionTime'},
  ],
};

const MsgUndelegate$json = const {
  '1': 'MsgUndelegate',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
    const {'1': 'amount', '3': 3, '4': 1, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'amount'},
  ],
  '7': const {},
};

const MsgUndelegateResponse$json = const {
  '1': 'MsgUndelegateResponse',
  '2': const [
    const {'1': 'completion_time', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'completionTime'},
  ],
};

