///
//  Generated code. Do not modify.
//  source: cosmos/slashing/v1beta1/genesis.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const GenesisState$json = const {
  '1': 'GenesisState',
  '2': const [
    const {'1': 'params', '3': 1, '4': 1, '5': 11, '6': '.cosmos.slashing.v1beta1.Params', '8': const {}, '10': 'params'},
    const {'1': 'signing_infos', '3': 2, '4': 3, '5': 11, '6': '.cosmos.slashing.v1beta1.SigningInfo', '8': const {}, '10': 'signingInfos'},
    const {'1': 'missed_blocks', '3': 3, '4': 3, '5': 11, '6': '.cosmos.slashing.v1beta1.ValidatorMissedBlocks', '8': const {}, '10': 'missedBlocks'},
  ],
};

const SigningInfo$json = const {
  '1': 'SigningInfo',
  '2': const [
    const {'1': 'address', '3': 1, '4': 1, '5': 9, '10': 'address'},
    const {'1': 'validator_signing_info', '3': 2, '4': 1, '5': 11, '6': '.cosmos.slashing.v1beta1.ValidatorSigningInfo', '8': const {}, '10': 'validatorSigningInfo'},
  ],
};

const ValidatorMissedBlocks$json = const {
  '1': 'ValidatorMissedBlocks',
  '2': const [
    const {'1': 'address', '3': 1, '4': 1, '5': 9, '10': 'address'},
    const {'1': 'missed_blocks', '3': 2, '4': 3, '5': 11, '6': '.cosmos.slashing.v1beta1.MissedBlock', '8': const {}, '10': 'missedBlocks'},
  ],
};

const MissedBlock$json = const {
  '1': 'MissedBlock',
  '2': const [
    const {'1': 'index', '3': 1, '4': 1, '5': 3, '10': 'index'},
    const {'1': 'missed', '3': 2, '4': 1, '5': 8, '10': 'missed'},
  ],
};

