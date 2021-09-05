///
//  Generated code. Do not modify.
//  source: cosmos/distribution/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const MsgSetWithdrawAddress$json = const {
  '1': 'MsgSetWithdrawAddress',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'withdraw_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'withdrawAddress'},
  ],
  '7': const {},
};

const MsgSetWithdrawAddressResponse$json = const {
  '1': 'MsgSetWithdrawAddressResponse',
};

const MsgWithdrawDelegatorReward$json = const {
  '1': 'MsgWithdrawDelegatorReward',
  '2': const [
    const {'1': 'delegator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'delegatorAddress'},
    const {'1': 'validator_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
  ],
  '7': const {},
};

const MsgWithdrawDelegatorRewardResponse$json = const {
  '1': 'MsgWithdrawDelegatorRewardResponse',
};

const MsgWithdrawValidatorCommission$json = const {
  '1': 'MsgWithdrawValidatorCommission',
  '2': const [
    const {'1': 'validator_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'validatorAddress'},
  ],
  '7': const {},
};

const MsgWithdrawValidatorCommissionResponse$json = const {
  '1': 'MsgWithdrawValidatorCommissionResponse',
};

const MsgFundCommunityPool$json = const {
  '1': 'MsgFundCommunityPool',
  '2': const [
    const {'1': 'amount', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'amount'},
    const {'1': 'depositor', '3': 2, '4': 1, '5': 9, '10': 'depositor'},
  ],
  '7': const {},
};

const MsgFundCommunityPoolResponse$json = const {
  '1': 'MsgFundCommunityPoolResponse',
};

