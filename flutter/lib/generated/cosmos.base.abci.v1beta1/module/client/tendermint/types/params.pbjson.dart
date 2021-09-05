///
//  Generated code. Do not modify.
//  source: tendermint/types/params.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const ConsensusParams$json = const {
  '1': 'ConsensusParams',
  '2': const [
    const {'1': 'block', '3': 1, '4': 1, '5': 11, '6': '.tendermint.types.BlockParams', '8': const {}, '10': 'block'},
    const {'1': 'evidence', '3': 2, '4': 1, '5': 11, '6': '.tendermint.types.EvidenceParams', '8': const {}, '10': 'evidence'},
    const {'1': 'validator', '3': 3, '4': 1, '5': 11, '6': '.tendermint.types.ValidatorParams', '8': const {}, '10': 'validator'},
    const {'1': 'version', '3': 4, '4': 1, '5': 11, '6': '.tendermint.types.VersionParams', '8': const {}, '10': 'version'},
  ],
};

const BlockParams$json = const {
  '1': 'BlockParams',
  '2': const [
    const {'1': 'max_bytes', '3': 1, '4': 1, '5': 3, '10': 'maxBytes'},
    const {'1': 'max_gas', '3': 2, '4': 1, '5': 3, '10': 'maxGas'},
    const {'1': 'time_iota_ms', '3': 3, '4': 1, '5': 3, '10': 'timeIotaMs'},
  ],
};

const EvidenceParams$json = const {
  '1': 'EvidenceParams',
  '2': const [
    const {'1': 'max_age_num_blocks', '3': 1, '4': 1, '5': 3, '10': 'maxAgeNumBlocks'},
    const {'1': 'max_age_duration', '3': 2, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'maxAgeDuration'},
    const {'1': 'max_bytes', '3': 3, '4': 1, '5': 3, '10': 'maxBytes'},
  ],
};

const ValidatorParams$json = const {
  '1': 'ValidatorParams',
  '2': const [
    const {'1': 'pub_key_types', '3': 1, '4': 3, '5': 9, '10': 'pubKeyTypes'},
  ],
  '7': const {},
};

const VersionParams$json = const {
  '1': 'VersionParams',
  '2': const [
    const {'1': 'app_version', '3': 1, '4': 1, '5': 4, '10': 'appVersion'},
  ],
  '7': const {},
};

const HashedParams$json = const {
  '1': 'HashedParams',
  '2': const [
    const {'1': 'block_max_bytes', '3': 1, '4': 1, '5': 3, '10': 'blockMaxBytes'},
    const {'1': 'block_max_gas', '3': 2, '4': 1, '5': 3, '10': 'blockMaxGas'},
  ],
};

