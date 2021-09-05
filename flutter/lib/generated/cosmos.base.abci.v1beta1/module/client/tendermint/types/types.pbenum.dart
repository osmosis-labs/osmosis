///
//  Generated code. Do not modify.
//  source: tendermint/types/types.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

// ignore_for_file: UNDEFINED_SHOWN_NAME
import 'dart:core' as $core;
import 'package:protobuf/protobuf.dart' as $pb;

class BlockIDFlag extends $pb.ProtobufEnum {
  static const BlockIDFlag BLOCK_ID_FLAG_UNKNOWN = BlockIDFlag._(0, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'BLOCK_ID_FLAG_UNKNOWN');
  static const BlockIDFlag BLOCK_ID_FLAG_ABSENT = BlockIDFlag._(1, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'BLOCK_ID_FLAG_ABSENT');
  static const BlockIDFlag BLOCK_ID_FLAG_COMMIT = BlockIDFlag._(2, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'BLOCK_ID_FLAG_COMMIT');
  static const BlockIDFlag BLOCK_ID_FLAG_NIL = BlockIDFlag._(3, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'BLOCK_ID_FLAG_NIL');

  static const $core.List<BlockIDFlag> values = <BlockIDFlag> [
    BLOCK_ID_FLAG_UNKNOWN,
    BLOCK_ID_FLAG_ABSENT,
    BLOCK_ID_FLAG_COMMIT,
    BLOCK_ID_FLAG_NIL,
  ];

  static final $core.Map<$core.int, BlockIDFlag> _byValue = $pb.ProtobufEnum.initByValue(values);
  static BlockIDFlag valueOf($core.int value) => _byValue[value];

  const BlockIDFlag._($core.int v, $core.String n) : super(v, n);
}

class SignedMsgType extends $pb.ProtobufEnum {
  static const SignedMsgType SIGNED_MSG_TYPE_UNKNOWN = SignedMsgType._(0, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'SIGNED_MSG_TYPE_UNKNOWN');
  static const SignedMsgType SIGNED_MSG_TYPE_PREVOTE = SignedMsgType._(1, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'SIGNED_MSG_TYPE_PREVOTE');
  static const SignedMsgType SIGNED_MSG_TYPE_PRECOMMIT = SignedMsgType._(2, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'SIGNED_MSG_TYPE_PRECOMMIT');
  static const SignedMsgType SIGNED_MSG_TYPE_PROPOSAL = SignedMsgType._(32, const $core.bool.fromEnvironment('protobuf.omit_enum_names') ? '' : 'SIGNED_MSG_TYPE_PROPOSAL');

  static const $core.List<SignedMsgType> values = <SignedMsgType> [
    SIGNED_MSG_TYPE_UNKNOWN,
    SIGNED_MSG_TYPE_PREVOTE,
    SIGNED_MSG_TYPE_PRECOMMIT,
    SIGNED_MSG_TYPE_PROPOSAL,
  ];

  static final $core.Map<$core.int, SignedMsgType> _byValue = $pb.ProtobufEnum.initByValue(values);
  static SignedMsgType valueOf($core.int value) => _byValue[value];

  const SignedMsgType._($core.int v, $core.String n) : super(v, n);
}

