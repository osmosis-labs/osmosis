///
//  Generated code. Do not modify.
//  source: cosmos/slashing/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const QueryParamsRequest$json = const {
  '1': 'QueryParamsRequest',
};

const QueryParamsResponse$json = const {
  '1': 'QueryParamsResponse',
  '2': const [
    const {'1': 'params', '3': 1, '4': 1, '5': 11, '6': '.cosmos.slashing.v1beta1.Params', '8': const {}, '10': 'params'},
  ],
};

const QuerySigningInfoRequest$json = const {
  '1': 'QuerySigningInfoRequest',
  '2': const [
    const {'1': 'cons_address', '3': 1, '4': 1, '5': 9, '10': 'consAddress'},
  ],
};

const QuerySigningInfoResponse$json = const {
  '1': 'QuerySigningInfoResponse',
  '2': const [
    const {'1': 'val_signing_info', '3': 1, '4': 1, '5': 11, '6': '.cosmos.slashing.v1beta1.ValidatorSigningInfo', '8': const {}, '10': 'valSigningInfo'},
  ],
};

const QuerySigningInfosRequest$json = const {
  '1': 'QuerySigningInfosRequest',
  '2': const [
    const {'1': 'pagination', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QuerySigningInfosResponse$json = const {
  '1': 'QuerySigningInfosResponse',
  '2': const [
    const {'1': 'info', '3': 1, '4': 3, '5': 11, '6': '.cosmos.slashing.v1beta1.ValidatorSigningInfo', '8': const {}, '10': 'info'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

