///
//  Generated code. Do not modify.
//  source: cosmos/evidence/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const QueryEvidenceRequest$json = const {
  '1': 'QueryEvidenceRequest',
  '2': const [
    const {'1': 'evidence_hash', '3': 1, '4': 1, '5': 12, '8': const {}, '10': 'evidenceHash'},
  ],
};

const QueryEvidenceResponse$json = const {
  '1': 'QueryEvidenceResponse',
  '2': const [
    const {'1': 'evidence', '3': 1, '4': 1, '5': 11, '6': '.google.protobuf.Any', '10': 'evidence'},
  ],
};

const QueryAllEvidenceRequest$json = const {
  '1': 'QueryAllEvidenceRequest',
  '2': const [
    const {'1': 'pagination', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageRequest', '10': 'pagination'},
  ],
};

const QueryAllEvidenceResponse$json = const {
  '1': 'QueryAllEvidenceResponse',
  '2': const [
    const {'1': 'evidence', '3': 1, '4': 3, '5': 11, '6': '.google.protobuf.Any', '10': 'evidence'},
    const {'1': 'pagination', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.query.v1beta1.PageResponse', '10': 'pagination'},
  ],
};

