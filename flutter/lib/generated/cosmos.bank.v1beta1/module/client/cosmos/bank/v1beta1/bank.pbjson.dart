///
//  Generated code. Do not modify.
//  source: cosmos/bank/v1beta1/bank.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const Params$json = const {
  '1': 'Params',
  '2': const [
    const {'1': 'send_enabled', '3': 1, '4': 3, '5': 11, '6': '.cosmos.bank.v1beta1.SendEnabled', '8': const {}, '10': 'sendEnabled'},
    const {'1': 'default_send_enabled', '3': 2, '4': 1, '5': 8, '8': const {}, '10': 'defaultSendEnabled'},
  ],
  '7': const {},
};

const SendEnabled$json = const {
  '1': 'SendEnabled',
  '2': const [
    const {'1': 'denom', '3': 1, '4': 1, '5': 9, '10': 'denom'},
    const {'1': 'enabled', '3': 2, '4': 1, '5': 8, '10': 'enabled'},
  ],
  '7': const {},
};

const Input$json = const {
  '1': 'Input',
  '2': const [
    const {'1': 'address', '3': 1, '4': 1, '5': 9, '10': 'address'},
    const {'1': 'coins', '3': 2, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
  '7': const {},
};

const Output$json = const {
  '1': 'Output',
  '2': const [
    const {'1': 'address', '3': 1, '4': 1, '5': 9, '10': 'address'},
    const {'1': 'coins', '3': 2, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
  ],
  '7': const {},
};

const Supply$json = const {
  '1': 'Supply',
  '2': const [
    const {'1': 'total', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'total'},
  ],
  '7': const {},
};

const DenomUnit$json = const {
  '1': 'DenomUnit',
  '2': const [
    const {'1': 'denom', '3': 1, '4': 1, '5': 9, '10': 'denom'},
    const {'1': 'exponent', '3': 2, '4': 1, '5': 13, '10': 'exponent'},
    const {'1': 'aliases', '3': 3, '4': 3, '5': 9, '10': 'aliases'},
  ],
};

const Metadata$json = const {
  '1': 'Metadata',
  '2': const [
    const {'1': 'description', '3': 1, '4': 1, '5': 9, '10': 'description'},
    const {'1': 'denom_units', '3': 2, '4': 3, '5': 11, '6': '.cosmos.bank.v1beta1.DenomUnit', '10': 'denomUnits'},
    const {'1': 'base', '3': 3, '4': 1, '5': 9, '10': 'base'},
    const {'1': 'display', '3': 4, '4': 1, '5': 9, '10': 'display'},
  ],
};

