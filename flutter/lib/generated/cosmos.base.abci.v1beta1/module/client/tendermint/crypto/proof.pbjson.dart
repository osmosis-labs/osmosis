///
//  Generated code. Do not modify.
//  source: tendermint/crypto/proof.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const Proof$json = const {
  '1': 'Proof',
  '2': const [
    const {'1': 'total', '3': 1, '4': 1, '5': 3, '10': 'total'},
    const {'1': 'index', '3': 2, '4': 1, '5': 3, '10': 'index'},
    const {'1': 'leaf_hash', '3': 3, '4': 1, '5': 12, '10': 'leafHash'},
    const {'1': 'aunts', '3': 4, '4': 3, '5': 12, '10': 'aunts'},
  ],
};

const ValueOp$json = const {
  '1': 'ValueOp',
  '2': const [
    const {'1': 'key', '3': 1, '4': 1, '5': 12, '10': 'key'},
    const {'1': 'proof', '3': 2, '4': 1, '5': 11, '6': '.tendermint.crypto.Proof', '10': 'proof'},
  ],
};

const DominoOp$json = const {
  '1': 'DominoOp',
  '2': const [
    const {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    const {'1': 'input', '3': 2, '4': 1, '5': 9, '10': 'input'},
    const {'1': 'output', '3': 3, '4': 1, '5': 9, '10': 'output'},
  ],
};

const ProofOp$json = const {
  '1': 'ProofOp',
  '2': const [
    const {'1': 'type', '3': 1, '4': 1, '5': 9, '10': 'type'},
    const {'1': 'key', '3': 2, '4': 1, '5': 12, '10': 'key'},
    const {'1': 'data', '3': 3, '4': 1, '5': 12, '10': 'data'},
  ],
};

const ProofOps$json = const {
  '1': 'ProofOps',
  '2': const [
    const {'1': 'ops', '3': 1, '4': 3, '5': 11, '6': '.tendermint.crypto.ProofOp', '8': const {}, '10': 'ops'},
  ],
};

