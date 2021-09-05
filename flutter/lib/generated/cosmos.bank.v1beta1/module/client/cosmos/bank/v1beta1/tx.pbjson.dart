///
//  Generated code. Do not modify.
//  source: cosmos/bank/v1beta1/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const MsgSend$json = const {
  '1': 'MsgSend',
  '2': const [
    const {'1': 'from_address', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'fromAddress'},
    const {'1': 'to_address', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'toAddress'},
    const {'1': 'amount', '3': 3, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'amount'},
  ],
  '7': const {},
};

const MsgSendResponse$json = const {
  '1': 'MsgSendResponse',
};

const MsgMultiSend$json = const {
  '1': 'MsgMultiSend',
  '2': const [
    const {'1': 'inputs', '3': 1, '4': 3, '5': 11, '6': '.cosmos.bank.v1beta1.Input', '8': const {}, '10': 'inputs'},
    const {'1': 'outputs', '3': 2, '4': 3, '5': 11, '6': '.cosmos.bank.v1beta1.Output', '8': const {}, '10': 'outputs'},
  ],
  '7': const {},
};

const MsgMultiSendResponse$json = const {
  '1': 'MsgMultiSendResponse',
};

