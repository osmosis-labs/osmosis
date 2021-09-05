///
//  Generated code. Do not modify.
//  source: ibc/applications/transfer/v1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const QueryDenomTraceRequest$json = const {
  '1': 'QueryDenomTraceRequest',
  '2': const [
    const {'1': 'hash', '3': 1, '4': 1, '5': 9, '10': 'hash'},
  ],
};

const QueryDenomTraceResponse$json = const {
  '1': 'QueryDenomTraceResponse',
  '2': const [
    const {'1': 'denom_trace', '3': 1, '4': 1, '5': 11, '6': '.ibc.applications.transfer.v1.DenomTrace', '10': 'denomTrace'},
  ],
};

const QueryDenomTracesRequest$json = const {
  '1': 'QueryDenomTracesRequest',
  '2': const [
    const {'1': 'pagination', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryDenomTracesResponse$json = const {
  '1': 'QueryDenomTracesResponse',
  '2': const [
    const {'1': 'denom_traces', '3': 1, '4': 3, '5': 11, '6': '.ibc.applications.transfer.v1.DenomTrace', '8': const {}, '10': 'denomTraces'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

const QueryParamsRequest$json = const {
  '1': 'QueryParamsRequest',
};

const QueryParamsResponse$json = const {
  '1': 'QueryParamsResponse',
  '2': const [
    const {'1': 'params', '3': 1, '4': 1, '5': 11, '6': '.ibc.applications.transfer.v1.Params', '10': 'params'},
  ],
};

