///
//  Generated code. Do not modify.
//  source: google/api/http.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const Http$json = const {
  '1': 'Http',
  '2': const [
    const {'1': 'rules', '3': 1, '4': 3, '5': 11, '6': '.google.api.HttpRule', '10': 'rules'},
    const {'1': 'fully_decode_reserved_expansion', '3': 2, '4': 1, '5': 8, '10': 'fullyDecodeReservedExpansion'},
  ],
};

const HttpRule$json = const {
  '1': 'HttpRule',
  '2': const [
    const {'1': 'selector', '3': 1, '4': 1, '5': 9, '10': 'selector'},
    const {'1': 'get', '3': 2, '4': 1, '5': 9, '9': 0, '10': 'get'},
    const {'1': 'put', '3': 3, '4': 1, '5': 9, '9': 0, '10': 'put'},
    const {'1': 'post', '3': 4, '4': 1, '5': 9, '9': 0, '10': 'post'},
    const {'1': 'delete', '3': 5, '4': 1, '5': 9, '9': 0, '10': 'delete'},
    const {'1': 'patch', '3': 6, '4': 1, '5': 9, '9': 0, '10': 'patch'},
    const {'1': 'custom', '3': 8, '4': 1, '5': 11, '6': '.google.api.CustomHttpPattern', '9': 0, '10': 'custom'},
    const {'1': 'body', '3': 7, '4': 1, '5': 9, '10': 'body'},
    const {'1': 'response_body', '3': 12, '4': 1, '5': 9, '10': 'responseBody'},
    const {'1': 'additional_bindings', '3': 11, '4': 3, '5': 11, '6': '.google.api.HttpRule', '10': 'additionalBindings'},
  ],
  '8': const [
    const {'1': 'pattern'},
  ],
};

const CustomHttpPattern$json = const {
  '1': 'CustomHttpPattern',
  '2': const [
    const {'1': 'kind', '3': 1, '4': 1, '5': 9, '10': 'kind'},
    const {'1': 'path', '3': 2, '4': 1, '5': 9, '10': 'path'},
  ],
};

