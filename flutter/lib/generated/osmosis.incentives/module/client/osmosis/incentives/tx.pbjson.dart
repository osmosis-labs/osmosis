///
//  Generated code. Do not modify.
//  source: osmosis/incentives/tx.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const MsgCreateGauge$json = const {
  '1': 'MsgCreateGauge',
  '2': const [
    const {'1': 'is_perpetual', '3': 1, '4': 1, '5': 8, '10': 'isPerpetual'},
    const {'1': 'owner', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'distribute_to', '3': 3, '4': 1, '5': 11, '6': '.osmosis.lockup.QueryCondition', '8': const {}, '10': 'distributeTo'},
    const {'1': 'coins', '3': 4, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'coins'},
    const {'1': 'start_time', '3': 5, '4': 1, '5': 11, '6': '.google.protobuf.Timestamp', '8': const {}, '10': 'startTime'},
    const {'1': 'num_epochs_paid_over', '3': 6, '4': 1, '5': 4, '10': 'numEpochsPaidOver'},
  ],
};

const MsgCreateGaugeResponse$json = const {
  '1': 'MsgCreateGaugeResponse',
};

const MsgAddToGauge$json = const {
  '1': 'MsgAddToGauge',
  '2': const [
    const {'1': 'owner', '3': 1, '4': 1, '5': 9, '8': const {}, '10': 'owner'},
    const {'1': 'gauge_id', '3': 2, '4': 1, '5': 4, '10': 'gaugeId'},
    const {'1': 'rewards', '3': 3, '4': 3, '5': 11, '6': '.cosmos.base.v1beta1.Coin', '8': const {}, '10': 'rewards'},
  ],
};

const MsgAddToGaugeResponse$json = const {
  '1': 'MsgAddToGaugeResponse',
};

