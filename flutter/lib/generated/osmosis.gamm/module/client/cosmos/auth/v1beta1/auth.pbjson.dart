///
//  Generated code. Do not modify.
//  source: cosmos/auth/v1beta1/auth.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const BaseAccount$json = const {
  '1': 'BaseAccount',
  '2': const [
    const {'1': 'address', '3': 1, '4': 1, '5': 9, '10': 'address'},
    const {'1': 'pub_key', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Any', '8': const {}, '10': 'pubKey'},
    const {'1': 'account_number', '3': 3, '4': 1, '5': 4, '8': const {}, '10': 'accountNumber'},
    const {'1': 'sequence', '3': 4, '4': 1, '5': 4, '10': 'sequence'},
  ],
  '7': const {},
};

const ModuleAccount$json = const {
  '1': 'ModuleAccount',
  '2': const [
    const {'1': 'base_account', '3': 1, '4': 1, '5': 11, '6': '.cosmos.auth.v1beta1.BaseAccount', '8': const {}, '10': 'baseAccount'},
    const {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    const {'1': 'permissions', '3': 3, '4': 3, '5': 9, '10': 'permissions'},
  ],
  '7': const {},
};

const Params$json = const {
  '1': 'Params',
  '2': const [
    const {'1': 'max_memo_characters', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'maxMemoCharacters'},
    const {'1': 'tx_sig_limit', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'txSigLimit'},
    const {'1': 'tx_size_cost_per_byte', '3': 3, '4': 1, '5': 4, '8': const {}, '10': 'txSizeCostPerByte'},
    const {'1': 'sig_verify_cost_ed25519', '3': 4, '4': 1, '5': 4, '8': const {}, '10': 'sigVerifyCostEd25519'},
    const {'1': 'sig_verify_cost_secp256k1', '3': 5, '4': 1, '5': 4, '8': const {}, '10': 'sigVerifyCostSecp256k1'},
  ],
  '7': const {},
};

