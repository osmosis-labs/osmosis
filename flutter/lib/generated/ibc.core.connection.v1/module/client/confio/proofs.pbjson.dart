///
//  Generated code. Do not modify.
//  source: confio/proofs.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const HashOp$json = const {
  '1': 'HashOp',
  '2': const [
    const {'1': 'NO_HASH', '2': 0},
    const {'1': 'SHA256', '2': 1},
    const {'1': 'SHA512', '2': 2},
    const {'1': 'KECCAK', '2': 3},
    const {'1': 'RIPEMD160', '2': 4},
    const {'1': 'BITCOIN', '2': 5},
  ],
};

const LengthOp$json = const {
  '1': 'LengthOp',
  '2': const [
    const {'1': 'NO_PREFIX', '2': 0},
    const {'1': 'VAR_PROTO', '2': 1},
    const {'1': 'VAR_RLP', '2': 2},
    const {'1': 'FIXED32_BIG', '2': 3},
    const {'1': 'FIXED32_LITTLE', '2': 4},
    const {'1': 'FIXED64_BIG', '2': 5},
    const {'1': 'FIXED64_LITTLE', '2': 6},
    const {'1': 'REQUIRE_32_BYTES', '2': 7},
    const {'1': 'REQUIRE_64_BYTES', '2': 8},
  ],
};

const ExistenceProof$json = const {
  '1': 'ExistenceProof',
  '2': const [
    const {'1': 'key', '3': 1, '4': 1, '5': 12, '10': 'key'},
    const {'1': 'value', '3': 2, '4': 1, '5': 12, '10': 'value'},
    const {'1': 'leaf', '3': 3, '4': 1, '5': 11, '6': '.ics23.LeafOp', '10': 'leaf'},
    const {'1': 'path', '3': 4, '4': 3, '5': 11, '6': '.ics23.InnerOp', '10': 'path'},
  ],
};

const NonExistenceProof$json = const {
  '1': 'NonExistenceProof',
  '2': const [
    const {'1': 'key', '3': 1, '4': 1, '5': 12, '10': 'key'},
    const {'1': 'left', '3': 2, '4': 1, '5': 11, '6': '.ics23.ExistenceProof', '10': 'left'},
    const {'1': 'right', '3': 3, '4': 1, '5': 11, '6': '.ics23.ExistenceProof', '10': 'right'},
  ],
};

const CommitmentProof$json = const {
  '1': 'CommitmentProof',
  '2': const [
    const {'1': 'exist', '3': 1, '4': 1, '5': 11, '6': '.ics23.ExistenceProof', '9': 0, '10': 'exist'},
    const {'1': 'nonexist', '3': 2, '4': 1, '5': 11, '6': '.ics23.NonExistenceProof', '9': 0, '10': 'nonexist'},
    const {'1': 'batch', '3': 3, '4': 1, '5': 11, '6': '.ics23.BatchProof', '9': 0, '10': 'batch'},
    const {'1': 'compressed', '3': 4, '4': 1, '5': 11, '6': '.ics23.CompressedBatchProof', '9': 0, '10': 'compressed'},
  ],
  '8': const [
    const {'1': 'proof'},
  ],
};

const LeafOp$json = const {
  '1': 'LeafOp',
  '2': const [
    const {'1': 'hash', '3': 1, '4': 1, '5': 14, '6': '.ics23.HashOp', '10': 'hash'},
    const {'1': 'prehash_key', '3': 2, '4': 1, '5': 14, '6': '.ics23.HashOp', '10': 'prehashKey'},
    const {'1': 'prehash_value', '3': 3, '4': 1, '5': 14, '6': '.ics23.HashOp', '10': 'prehashValue'},
    const {'1': 'length', '3': 4, '4': 1, '5': 14, '6': '.ics23.LengthOp', '10': 'length'},
    const {'1': 'prefix', '3': 5, '4': 1, '5': 12, '10': 'prefix'},
  ],
};

const InnerOp$json = const {
  '1': 'InnerOp',
  '2': const [
    const {'1': 'hash', '3': 1, '4': 1, '5': 14, '6': '.ics23.HashOp', '10': 'hash'},
    const {'1': 'prefix', '3': 2, '4': 1, '5': 12, '10': 'prefix'},
    const {'1': 'suffix', '3': 3, '4': 1, '5': 12, '10': 'suffix'},
  ],
};

const ProofSpec$json = const {
  '1': 'ProofSpec',
  '2': const [
    const {'1': 'leaf_spec', '3': 1, '4': 1, '5': 11, '6': '.ics23.LeafOp', '10': 'leafSpec'},
    const {'1': 'inner_spec', '3': 2, '4': 1, '5': 11, '6': '.ics23.InnerSpec', '10': 'innerSpec'},
    const {'1': 'max_depth', '3': 3, '4': 1, '5': 5, '10': 'maxDepth'},
    const {'1': 'min_depth', '3': 4, '4': 1, '5': 5, '10': 'minDepth'},
  ],
};

const InnerSpec$json = const {
  '1': 'InnerSpec',
  '2': const [
    const {'1': 'child_order', '3': 1, '4': 3, '5': 5, '10': 'childOrder'},
    const {'1': 'child_size', '3': 2, '4': 1, '5': 5, '10': 'childSize'},
    const {'1': 'min_prefix_length', '3': 3, '4': 1, '5': 5, '10': 'minPrefixLength'},
    const {'1': 'max_prefix_length', '3': 4, '4': 1, '5': 5, '10': 'maxPrefixLength'},
    const {'1': 'empty_child', '3': 5, '4': 1, '5': 12, '10': 'emptyChild'},
    const {'1': 'hash', '3': 6, '4': 1, '5': 14, '6': '.ics23.HashOp', '10': 'hash'},
  ],
};

const BatchProof$json = const {
  '1': 'BatchProof',
  '2': const [
    const {'1': 'entries', '3': 1, '4': 3, '5': 11, '6': '.ics23.BatchEntry', '10': 'entries'},
  ],
};

const BatchEntry$json = const {
  '1': 'BatchEntry',
  '2': const [
    const {'1': 'exist', '3': 1, '4': 1, '5': 11, '6': '.ics23.ExistenceProof', '9': 0, '10': 'exist'},
    const {'1': 'nonexist', '3': 2, '4': 1, '5': 11, '6': '.ics23.NonExistenceProof', '9': 0, '10': 'nonexist'},
  ],
  '8': const [
    const {'1': 'proof'},
  ],
};

const CompressedBatchProof$json = const {
  '1': 'CompressedBatchProof',
  '2': const [
    const {'1': 'entries', '3': 1, '4': 3, '5': 11, '6': '.ics23.CompressedBatchEntry', '10': 'entries'},
    const {'1': 'lookup_inners', '3': 2, '4': 3, '5': 11, '6': '.ics23.InnerOp', '10': 'lookupInners'},
  ],
};

const CompressedBatchEntry$json = const {
  '1': 'CompressedBatchEntry',
  '2': const [
    const {'1': 'exist', '3': 1, '4': 1, '5': 11, '6': '.ics23.CompressedExistenceProof', '9': 0, '10': 'exist'},
    const {'1': 'nonexist', '3': 2, '4': 1, '5': 11, '6': '.ics23.CompressedNonExistenceProof', '9': 0, '10': 'nonexist'},
  ],
  '8': const [
    const {'1': 'proof'},
  ],
};

const CompressedExistenceProof$json = const {
  '1': 'CompressedExistenceProof',
  '2': const [
    const {'1': 'key', '3': 1, '4': 1, '5': 12, '10': 'key'},
    const {'1': 'value', '3': 2, '4': 1, '5': 12, '10': 'value'},
    const {'1': 'leaf', '3': 3, '4': 1, '5': 11, '6': '.ics23.LeafOp', '10': 'leaf'},
    const {'1': 'path', '3': 4, '4': 3, '5': 5, '10': 'path'},
  ],
};

const CompressedNonExistenceProof$json = const {
  '1': 'CompressedNonExistenceProof',
  '2': const [
    const {'1': 'key', '3': 1, '4': 1, '5': 12, '10': 'key'},
    const {'1': 'left', '3': 2, '4': 1, '5': 11, '6': '.ics23.CompressedExistenceProof', '10': 'left'},
    const {'1': 'right', '3': 3, '4': 1, '5': 11, '6': '.ics23.CompressedExistenceProof', '10': 'right'},
  ],
};

