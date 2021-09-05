///
//  Generated code. Do not modify.
//  source: osmosis/lockup/lock.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

// ignore_for_file: UNDEFINED_SHOWN_NAME
import 'dart:core' as $core;
import 'package:protobuf/protobuf.dart' as $pb;

class LockQueryType extends $pb.ProtobufEnum {
  static const LockQueryType ByDuration = LockQueryType._(0, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'ByDuration');
  static const LockQueryType ByTime = LockQueryType._(1, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'ByTime');

  static const $core.List<LockQueryType> values = <LockQueryType> [
    ByDuration,
    ByTime,
  ];

  static final $core.Map<$core.int, LockQueryType> _byValue = $pb.ProtobufEnum.initByValue(values);
  static LockQueryType valueOf($core.int value) => _byValue[value];

  const LockQueryType._($core.int v, $core.String n) : super(v, n);
}

