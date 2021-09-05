///
//  Generated code. Do not modify.
//  source: cosmos/slashing/v1beta1/slashing.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const ValidatorSigningInfo$json = const {
  '1': 'ValidatorSigningInfo',
  '2': const [
    const {'1': 'address', '3': 1, '4': 1, '5': 9, '10': 'address'},
    const {'1': 'start_height', '3': 2, '4': 1, '5': 3, '8': const {}, '10': 'startHeight'},
    const {'1': 'index_offset', '3': 3, '4': 1, '5': 3, '8': const {}, '10': 'indexOffset'},
    const {'1': 'jailed_until', '3': 4, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'jailedUntil'},
    const {'1': 'tombstoned', '3': 5, '4': 1, '5': 8, '10': 'tombstoned'},
    const {'1': 'missed_blocks_counter', '3': 6, '4': 1, '5': 3, '8': const {}, '10': 'missedBlocksCounter'},
  ],
  '7': const {},
};

const Params$json = const {
  '1': 'Params',
  '2': const [
    const {'1': 'signed_blocks_window', '3': 1, '4': 1, '5': 3, '8': const {}, '10': 'signedBlocksWindow'},
    const {'1': 'min_signed_per_window', '3': 2, '4': 1, '5': 12, '8': const {}, '10': 'minSignedPerWindow'},
    const {'1': 'downtime_jail_duration', '3': 3, '4': 1, '5': 11, '6': '.google.protobuf.Duration', '8': const {}, '10': 'downtimeJailDuration'},
    const {'1': 'slash_fraction_double_sign', '3': 4, '4': 1, '5': 12, '8': const {}, '10': 'slashFractionDoubleSign'},
    const {'1': 'slash_fraction_downtime', '3': 5, '4': 1, '5': 12, '8': const {}, '10': 'slashFractionDowntime'},
  ],
};

