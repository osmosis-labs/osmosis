///
//  Generated code. Do not modify.
//  source: confio/proofs.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

// ignore_for_file: UNDEFINED_SHOWN_NAME
import 'dart:core' as $core;
import 'package:protobuf/protobuf.dart' as $pb;

class HashOp extends $pb.ProtobufEnum {
  static const HashOp NO_HASH = HashOp._(0, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'NO_HASH');
  static const HashOp SHA256 = HashOp._(1, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'SHA256');
  static const HashOp SHA512 = HashOp._(2, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'SHA512');
  static const HashOp KECCAK = HashOp._(3, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'KECCAK');
  static const HashOp RIPEMD160 = HashOp._(4, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'RIPEMD160');
  static const HashOp BITCOIN = HashOp._(5, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'BITCOIN');

  static const $core.List<HashOp> values = <HashOp> [
    NO_HASH,
    SHA256,
    SHA512,
    KECCAK,
    RIPEMD160,
    BITCOIN,
  ];

  static final $core.Map<$core.int, HashOp> _byValue = $pb.ProtobufEnum.initByValue(values);
  static HashOp valueOf($core.int value) => _byValue[value];

  const HashOp._($core.int v, $core.String n) : super(v, n);
}

class LengthOp extends $pb.ProtobufEnum {
  static const LengthOp NO_PREFIX = LengthOp._(0, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'NO_PREFIX');
  static const LengthOp VAR_PROTO = LengthOp._(1, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'VAR_PROTO');
  static const LengthOp VAR_RLP = LengthOp._(2, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'VAR_RLP');
  static const LengthOp FIXED32_BIG = LengthOp._(3, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'FIXED32_BIG');
  static const LengthOp FIXED32_LITTLE = LengthOp._(4, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'FIXED32_LITTLE');
  static const LengthOp FIXED64_BIG = LengthOp._(5, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'FIXED64_BIG');
  static const LengthOp FIXED64_LITTLE = LengthOp._(6, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'FIXED64_LITTLE');
  static const LengthOp REQUIRE_32_BYTES = LengthOp._(7, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'REQUIRE_32_BYTES');
  static const LengthOp REQUIRE_64_BYTES = LengthOp._(8, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'REQUIRE_64_BYTES');

  static const $core.List<LengthOp> values = <LengthOp> [
    NO_PREFIX,
    VAR_PROTO,
    VAR_RLP,
    FIXED32_BIG,
    FIXED32_LITTLE,
    FIXED64_BIG,
    FIXED64_LITTLE,
    REQUIRE_32_BYTES,
    REQUIRE_64_BYTES,
  ];

  static final $core.Map<$core.int, LengthOp> _byValue = $pb.ProtobufEnum.initByValue(values);
  static LengthOp valueOf($core.int value) => _byValue[value];

  const LengthOp._($core.int v, $core.String n) : super(v, n);
}

