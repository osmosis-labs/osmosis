///
//  Generated code. Do not modify.
//  source: ibc/core/client/v1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../../../google/protobuf/any.pb.dart' as $2;
import 'client.pb.dart' as $3;
import '../../../../cosmos/base/query/v1beta1/pagination.pb.dart' as $5;

class QueryClientStateRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryClientStateRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..hasRequiredFields = false
  ;

  QueryClientStateRequest._() : super();
  factory QueryClientStateRequest() => create();
  factory QueryClientStateRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryClientStateRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryClientStateRequest clone() => QueryClientStateRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryClientStateRequest copyWith(void Function(QueryClientStateRequest) updates) => super.copyWith((message) => updates(message as QueryClientStateRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryClientStateRequest create() => QueryClientStateRequest._();
  QueryClientStateRequest createEmptyInstance() => create();
  static $pb.PbList<QueryClientStateRequest> createRepeated() => $pb.PbList<QueryClientStateRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryClientStateRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryClientStateRequest>(create);
  static QueryClientStateRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);
}

class QueryClientStateResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryClientStateResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOM<$2.Any>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientState', subBuilder: $2.Any.create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proof', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryClientStateResponse._() : super();
  factory QueryClientStateResponse() => create();
  factory QueryClientStateResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryClientStateResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryClientStateResponse clone() => QueryClientStateResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryClientStateResponse copyWith(void Function(QueryClientStateResponse) updates) => super.copyWith((message) => updates(message as QueryClientStateResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryClientStateResponse create() => QueryClientStateResponse._();
  QueryClientStateResponse createEmptyInstance() => create();
  static $pb.PbList<QueryClientStateResponse> createRepeated() => $pb.PbList<QueryClientStateResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryClientStateResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryClientStateResponse>(create);
  static QueryClientStateResponse _defaultInstance;

  @$pb.TagNumber(1)
  $2.Any get clientState => $_getN(0);
  @$pb.TagNumber(1)
  set clientState($2.Any v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientState() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientState() => clearField(1);
  @$pb.TagNumber(1)
  $2.Any ensureClientState() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<$core.int> get proof => $_getN(1);
  @$pb.TagNumber(2)
  set proof($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasProof() => $_has(1);
  @$pb.TagNumber(2)
  void clearProof() => clearField(2);

  @$pb.TagNumber(3)
  $3.Height get proofHeight => $_getN(2);
  @$pb.TagNumber(3)
  set proofHeight($3.Height v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasProofHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearProofHeight() => clearField(3);
  @$pb.TagNumber(3)
  $3.Height ensureProofHeight() => $_ensure(2);
}

class QueryClientStatesRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryClientStatesRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOM<$5.PageRequest>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $5.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryClientStatesRequest._() : super();
  factory QueryClientStatesRequest() => create();
  factory QueryClientStatesRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryClientStatesRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryClientStatesRequest clone() => QueryClientStatesRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryClientStatesRequest copyWith(void Function(QueryClientStatesRequest) updates) => super.copyWith((message) => updates(message as QueryClientStatesRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryClientStatesRequest create() => QueryClientStatesRequest._();
  QueryClientStatesRequest createEmptyInstance() => create();
  static $pb.PbList<QueryClientStatesRequest> createRepeated() => $pb.PbList<QueryClientStatesRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryClientStatesRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryClientStatesRequest>(create);
  static QueryClientStatesRequest _defaultInstance;

  @$pb.TagNumber(1)
  $5.PageRequest get pagination => $_getN(0);
  @$pb.TagNumber(1)
  set pagination($5.PageRequest v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPagination() => $_has(0);
  @$pb.TagNumber(1)
  void clearPagination() => clearField(1);
  @$pb.TagNumber(1)
  $5.PageRequest ensurePagination() => $_ensure(0);
}

class QueryClientStatesResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryClientStatesResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..pc<$3.IdentifiedClientState>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientStates', $pb.PbFieldType.PM, subBuilder: $3.IdentifiedClientState.create)
    ..aOM<$5.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $5.PageResponse.create)
    ..hasRequiredFields = false
  ;

  QueryClientStatesResponse._() : super();
  factory QueryClientStatesResponse() => create();
  factory QueryClientStatesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryClientStatesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryClientStatesResponse clone() => QueryClientStatesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryClientStatesResponse copyWith(void Function(QueryClientStatesResponse) updates) => super.copyWith((message) => updates(message as QueryClientStatesResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryClientStatesResponse create() => QueryClientStatesResponse._();
  QueryClientStatesResponse createEmptyInstance() => create();
  static $pb.PbList<QueryClientStatesResponse> createRepeated() => $pb.PbList<QueryClientStatesResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryClientStatesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryClientStatesResponse>(create);
  static QueryClientStatesResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$3.IdentifiedClientState> get clientStates => $_getList(0);

  @$pb.TagNumber(2)
  $5.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($5.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $5.PageResponse ensurePagination() => $_ensure(1);
}

class QueryConsensusStateRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryConsensusStateRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'revisionNumber', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'revisionHeight', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOB(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'latestHeight')
    ..hasRequiredFields = false
  ;

  QueryConsensusStateRequest._() : super();
  factory QueryConsensusStateRequest() => create();
  factory QueryConsensusStateRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryConsensusStateRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryConsensusStateRequest clone() => QueryConsensusStateRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryConsensusStateRequest copyWith(void Function(QueryConsensusStateRequest) updates) => super.copyWith((message) => updates(message as QueryConsensusStateRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryConsensusStateRequest create() => QueryConsensusStateRequest._();
  QueryConsensusStateRequest createEmptyInstance() => create();
  static $pb.PbList<QueryConsensusStateRequest> createRepeated() => $pb.PbList<QueryConsensusStateRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryConsensusStateRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryConsensusStateRequest>(create);
  static QueryConsensusStateRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get revisionNumber => $_getI64(1);
  @$pb.TagNumber(2)
  set revisionNumber($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasRevisionNumber() => $_has(1);
  @$pb.TagNumber(2)
  void clearRevisionNumber() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get revisionHeight => $_getI64(2);
  @$pb.TagNumber(3)
  set revisionHeight($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasRevisionHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearRevisionHeight() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get latestHeight => $_getBF(3);
  @$pb.TagNumber(4)
  set latestHeight($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasLatestHeight() => $_has(3);
  @$pb.TagNumber(4)
  void clearLatestHeight() => clearField(4);
}

class QueryConsensusStateResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryConsensusStateResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOM<$2.Any>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusState', subBuilder: $2.Any.create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proof', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryConsensusStateResponse._() : super();
  factory QueryConsensusStateResponse() => create();
  factory QueryConsensusStateResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryConsensusStateResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryConsensusStateResponse clone() => QueryConsensusStateResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryConsensusStateResponse copyWith(void Function(QueryConsensusStateResponse) updates) => super.copyWith((message) => updates(message as QueryConsensusStateResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryConsensusStateResponse create() => QueryConsensusStateResponse._();
  QueryConsensusStateResponse createEmptyInstance() => create();
  static $pb.PbList<QueryConsensusStateResponse> createRepeated() => $pb.PbList<QueryConsensusStateResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryConsensusStateResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryConsensusStateResponse>(create);
  static QueryConsensusStateResponse _defaultInstance;

  @$pb.TagNumber(1)
  $2.Any get consensusState => $_getN(0);
  @$pb.TagNumber(1)
  set consensusState($2.Any v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasConsensusState() => $_has(0);
  @$pb.TagNumber(1)
  void clearConsensusState() => clearField(1);
  @$pb.TagNumber(1)
  $2.Any ensureConsensusState() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.List<$core.int> get proof => $_getN(1);
  @$pb.TagNumber(2)
  set proof($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasProof() => $_has(1);
  @$pb.TagNumber(2)
  void clearProof() => clearField(2);

  @$pb.TagNumber(3)
  $3.Height get proofHeight => $_getN(2);
  @$pb.TagNumber(3)
  set proofHeight($3.Height v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasProofHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearProofHeight() => clearField(3);
  @$pb.TagNumber(3)
  $3.Height ensureProofHeight() => $_ensure(2);
}

class QueryConsensusStatesRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryConsensusStatesRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..aOM<$5.PageRequest>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $5.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryConsensusStatesRequest._() : super();
  factory QueryConsensusStatesRequest() => create();
  factory QueryConsensusStatesRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryConsensusStatesRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryConsensusStatesRequest clone() => QueryConsensusStatesRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryConsensusStatesRequest copyWith(void Function(QueryConsensusStatesRequest) updates) => super.copyWith((message) => updates(message as QueryConsensusStatesRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryConsensusStatesRequest create() => QueryConsensusStatesRequest._();
  QueryConsensusStatesRequest createEmptyInstance() => create();
  static $pb.PbList<QueryConsensusStatesRequest> createRepeated() => $pb.PbList<QueryConsensusStatesRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryConsensusStatesRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryConsensusStatesRequest>(create);
  static QueryConsensusStatesRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $5.PageRequest get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($5.PageRequest v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $5.PageRequest ensurePagination() => $_ensure(1);
}

class QueryConsensusStatesResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryConsensusStatesResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..pc<$3.ConsensusStateWithHeight>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusStates', $pb.PbFieldType.PM, subBuilder: $3.ConsensusStateWithHeight.create)
    ..aOM<$5.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $5.PageResponse.create)
    ..hasRequiredFields = false
  ;

  QueryConsensusStatesResponse._() : super();
  factory QueryConsensusStatesResponse() => create();
  factory QueryConsensusStatesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryConsensusStatesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryConsensusStatesResponse clone() => QueryConsensusStatesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryConsensusStatesResponse copyWith(void Function(QueryConsensusStatesResponse) updates) => super.copyWith((message) => updates(message as QueryConsensusStatesResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryConsensusStatesResponse create() => QueryConsensusStatesResponse._();
  QueryConsensusStatesResponse createEmptyInstance() => create();
  static $pb.PbList<QueryConsensusStatesResponse> createRepeated() => $pb.PbList<QueryConsensusStatesResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryConsensusStatesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryConsensusStatesResponse>(create);
  static QueryConsensusStatesResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$3.ConsensusStateWithHeight> get consensusStates => $_getList(0);

  @$pb.TagNumber(2)
  $5.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($5.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $5.PageResponse ensurePagination() => $_ensure(1);
}

class QueryClientParamsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryClientParamsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  QueryClientParamsRequest._() : super();
  factory QueryClientParamsRequest() => create();
  factory QueryClientParamsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryClientParamsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryClientParamsRequest clone() => QueryClientParamsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryClientParamsRequest copyWith(void Function(QueryClientParamsRequest) updates) => super.copyWith((message) => updates(message as QueryClientParamsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryClientParamsRequest create() => QueryClientParamsRequest._();
  QueryClientParamsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryClientParamsRequest> createRepeated() => $pb.PbList<QueryClientParamsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryClientParamsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryClientParamsRequest>(create);
  static QueryClientParamsRequest _defaultInstance;
}

class QueryClientParamsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryClientParamsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.client.v1'), createEmptyInstance: create)
    ..aOM<$3.Params>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'params', subBuilder: $3.Params.create)
    ..hasRequiredFields = false
  ;

  QueryClientParamsResponse._() : super();
  factory QueryClientParamsResponse() => create();
  factory QueryClientParamsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryClientParamsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryClientParamsResponse clone() => QueryClientParamsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryClientParamsResponse copyWith(void Function(QueryClientParamsResponse) updates) => super.copyWith((message) => updates(message as QueryClientParamsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryClientParamsResponse create() => QueryClientParamsResponse._();
  QueryClientParamsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryClientParamsResponse> createRepeated() => $pb.PbList<QueryClientParamsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryClientParamsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryClientParamsResponse>(create);
  static QueryClientParamsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $3.Params get params => $_getN(0);
  @$pb.TagNumber(1)
  set params($3.Params v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasParams() => $_has(0);
  @$pb.TagNumber(1)
  void clearParams() => clearField(1);
  @$pb.TagNumber(1)
  $3.Params ensureParams() => $_ensure(0);
}

