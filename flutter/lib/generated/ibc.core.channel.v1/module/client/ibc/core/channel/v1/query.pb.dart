///
//  Generated code. Do not modify.
//  source: ibc/core/channel/v1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'channel.pb.dart' as $4;
import '../../client/v1/client.pb.dart' as $3;
import '../../../../cosmos/base/query/v1beta1/pagination.pb.dart' as $6;
import '../../../../google/protobuf/any.pb.dart' as $2;

class QueryChannelRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryChannelRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..hasRequiredFields = false
  ;

  QueryChannelRequest._() : super();
  factory QueryChannelRequest() => create();
  factory QueryChannelRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryChannelRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryChannelRequest clone() => QueryChannelRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryChannelRequest copyWith(void Function(QueryChannelRequest) updates) => super.copyWith((message) => updates(message as QueryChannelRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryChannelRequest create() => QueryChannelRequest._();
  QueryChannelRequest createEmptyInstance() => create();
  static $pb.PbList<QueryChannelRequest> createRepeated() => $pb.PbList<QueryChannelRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryChannelRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryChannelRequest>(create);
  static QueryChannelRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);
}

class QueryChannelResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryChannelResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOM<$4.Channel>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channel', subBuilder: $4.Channel.create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proof', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryChannelResponse._() : super();
  factory QueryChannelResponse() => create();
  factory QueryChannelResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryChannelResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryChannelResponse clone() => QueryChannelResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryChannelResponse copyWith(void Function(QueryChannelResponse) updates) => super.copyWith((message) => updates(message as QueryChannelResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryChannelResponse create() => QueryChannelResponse._();
  QueryChannelResponse createEmptyInstance() => create();
  static $pb.PbList<QueryChannelResponse> createRepeated() => $pb.PbList<QueryChannelResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryChannelResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryChannelResponse>(create);
  static QueryChannelResponse _defaultInstance;

  @$pb.TagNumber(1)
  $4.Channel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel($4.Channel v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => clearField(1);
  @$pb.TagNumber(1)
  $4.Channel ensureChannel() => $_ensure(0);

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

class QueryChannelsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryChannelsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOM<$6.PageRequest>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $6.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryChannelsRequest._() : super();
  factory QueryChannelsRequest() => create();
  factory QueryChannelsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryChannelsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryChannelsRequest clone() => QueryChannelsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryChannelsRequest copyWith(void Function(QueryChannelsRequest) updates) => super.copyWith((message) => updates(message as QueryChannelsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryChannelsRequest create() => QueryChannelsRequest._();
  QueryChannelsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryChannelsRequest> createRepeated() => $pb.PbList<QueryChannelsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryChannelsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryChannelsRequest>(create);
  static QueryChannelsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $6.PageRequest get pagination => $_getN(0);
  @$pb.TagNumber(1)
  set pagination($6.PageRequest v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPagination() => $_has(0);
  @$pb.TagNumber(1)
  void clearPagination() => clearField(1);
  @$pb.TagNumber(1)
  $6.PageRequest ensurePagination() => $_ensure(0);
}

class QueryChannelsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryChannelsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..pc<$4.IdentifiedChannel>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channels', $pb.PbFieldType.PM, subBuilder: $4.IdentifiedChannel.create)
    ..aOM<$6.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $6.PageResponse.create)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryChannelsResponse._() : super();
  factory QueryChannelsResponse() => create();
  factory QueryChannelsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryChannelsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryChannelsResponse clone() => QueryChannelsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryChannelsResponse copyWith(void Function(QueryChannelsResponse) updates) => super.copyWith((message) => updates(message as QueryChannelsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryChannelsResponse create() => QueryChannelsResponse._();
  QueryChannelsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryChannelsResponse> createRepeated() => $pb.PbList<QueryChannelsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryChannelsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryChannelsResponse>(create);
  static QueryChannelsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.IdentifiedChannel> get channels => $_getList(0);

  @$pb.TagNumber(2)
  $6.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($6.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $6.PageResponse ensurePagination() => $_ensure(1);

  @$pb.TagNumber(3)
  $3.Height get height => $_getN(2);
  @$pb.TagNumber(3)
  set height($3.Height v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearHeight() => clearField(3);
  @$pb.TagNumber(3)
  $3.Height ensureHeight() => $_ensure(2);
}

class QueryConnectionChannelsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryConnectionChannelsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'connection')
    ..aOM<$6.PageRequest>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $6.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryConnectionChannelsRequest._() : super();
  factory QueryConnectionChannelsRequest() => create();
  factory QueryConnectionChannelsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryConnectionChannelsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryConnectionChannelsRequest clone() => QueryConnectionChannelsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryConnectionChannelsRequest copyWith(void Function(QueryConnectionChannelsRequest) updates) => super.copyWith((message) => updates(message as QueryConnectionChannelsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryConnectionChannelsRequest create() => QueryConnectionChannelsRequest._();
  QueryConnectionChannelsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryConnectionChannelsRequest> createRepeated() => $pb.PbList<QueryConnectionChannelsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryConnectionChannelsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryConnectionChannelsRequest>(create);
  static QueryConnectionChannelsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get connection => $_getSZ(0);
  @$pb.TagNumber(1)
  set connection($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasConnection() => $_has(0);
  @$pb.TagNumber(1)
  void clearConnection() => clearField(1);

  @$pb.TagNumber(2)
  $6.PageRequest get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($6.PageRequest v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $6.PageRequest ensurePagination() => $_ensure(1);
}

class QueryConnectionChannelsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryConnectionChannelsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..pc<$4.IdentifiedChannel>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channels', $pb.PbFieldType.PM, subBuilder: $4.IdentifiedChannel.create)
    ..aOM<$6.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $6.PageResponse.create)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryConnectionChannelsResponse._() : super();
  factory QueryConnectionChannelsResponse() => create();
  factory QueryConnectionChannelsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryConnectionChannelsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryConnectionChannelsResponse clone() => QueryConnectionChannelsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryConnectionChannelsResponse copyWith(void Function(QueryConnectionChannelsResponse) updates) => super.copyWith((message) => updates(message as QueryConnectionChannelsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryConnectionChannelsResponse create() => QueryConnectionChannelsResponse._();
  QueryConnectionChannelsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryConnectionChannelsResponse> createRepeated() => $pb.PbList<QueryConnectionChannelsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryConnectionChannelsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryConnectionChannelsResponse>(create);
  static QueryConnectionChannelsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.IdentifiedChannel> get channels => $_getList(0);

  @$pb.TagNumber(2)
  $6.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($6.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $6.PageResponse ensurePagination() => $_ensure(1);

  @$pb.TagNumber(3)
  $3.Height get height => $_getN(2);
  @$pb.TagNumber(3)
  set height($3.Height v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearHeight() => clearField(3);
  @$pb.TagNumber(3)
  $3.Height ensureHeight() => $_ensure(2);
}

class QueryChannelClientStateRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryChannelClientStateRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..hasRequiredFields = false
  ;

  QueryChannelClientStateRequest._() : super();
  factory QueryChannelClientStateRequest() => create();
  factory QueryChannelClientStateRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryChannelClientStateRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryChannelClientStateRequest clone() => QueryChannelClientStateRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryChannelClientStateRequest copyWith(void Function(QueryChannelClientStateRequest) updates) => super.copyWith((message) => updates(message as QueryChannelClientStateRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryChannelClientStateRequest create() => QueryChannelClientStateRequest._();
  QueryChannelClientStateRequest createEmptyInstance() => create();
  static $pb.PbList<QueryChannelClientStateRequest> createRepeated() => $pb.PbList<QueryChannelClientStateRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryChannelClientStateRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryChannelClientStateRequest>(create);
  static QueryChannelClientStateRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);
}

class QueryChannelClientStateResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryChannelClientStateResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOM<$3.IdentifiedClientState>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'identifiedClientState', subBuilder: $3.IdentifiedClientState.create)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proof', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryChannelClientStateResponse._() : super();
  factory QueryChannelClientStateResponse() => create();
  factory QueryChannelClientStateResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryChannelClientStateResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryChannelClientStateResponse clone() => QueryChannelClientStateResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryChannelClientStateResponse copyWith(void Function(QueryChannelClientStateResponse) updates) => super.copyWith((message) => updates(message as QueryChannelClientStateResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryChannelClientStateResponse create() => QueryChannelClientStateResponse._();
  QueryChannelClientStateResponse createEmptyInstance() => create();
  static $pb.PbList<QueryChannelClientStateResponse> createRepeated() => $pb.PbList<QueryChannelClientStateResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryChannelClientStateResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryChannelClientStateResponse>(create);
  static QueryChannelClientStateResponse _defaultInstance;

  @$pb.TagNumber(1)
  $3.IdentifiedClientState get identifiedClientState => $_getN(0);
  @$pb.TagNumber(1)
  set identifiedClientState($3.IdentifiedClientState v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasIdentifiedClientState() => $_has(0);
  @$pb.TagNumber(1)
  void clearIdentifiedClientState() => clearField(1);
  @$pb.TagNumber(1)
  $3.IdentifiedClientState ensureIdentifiedClientState() => $_ensure(0);

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

class QueryChannelConsensusStateRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryChannelConsensusStateRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'revisionNumber', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'revisionHeight', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryChannelConsensusStateRequest._() : super();
  factory QueryChannelConsensusStateRequest() => create();
  factory QueryChannelConsensusStateRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryChannelConsensusStateRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryChannelConsensusStateRequest clone() => QueryChannelConsensusStateRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryChannelConsensusStateRequest copyWith(void Function(QueryChannelConsensusStateRequest) updates) => super.copyWith((message) => updates(message as QueryChannelConsensusStateRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryChannelConsensusStateRequest create() => QueryChannelConsensusStateRequest._();
  QueryChannelConsensusStateRequest createEmptyInstance() => create();
  static $pb.PbList<QueryChannelConsensusStateRequest> createRepeated() => $pb.PbList<QueryChannelConsensusStateRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryChannelConsensusStateRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryChannelConsensusStateRequest>(create);
  static QueryChannelConsensusStateRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get revisionNumber => $_getI64(2);
  @$pb.TagNumber(3)
  set revisionNumber($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasRevisionNumber() => $_has(2);
  @$pb.TagNumber(3)
  void clearRevisionNumber() => clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get revisionHeight => $_getI64(3);
  @$pb.TagNumber(4)
  set revisionHeight($fixnum.Int64 v) { $_setInt64(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasRevisionHeight() => $_has(3);
  @$pb.TagNumber(4)
  void clearRevisionHeight() => clearField(4);
}

class QueryChannelConsensusStateResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryChannelConsensusStateResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOM<$2.Any>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'consensusState', subBuilder: $2.Any.create)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'clientId')
    ..a<$core.List<$core.int>>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proof', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryChannelConsensusStateResponse._() : super();
  factory QueryChannelConsensusStateResponse() => create();
  factory QueryChannelConsensusStateResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryChannelConsensusStateResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryChannelConsensusStateResponse clone() => QueryChannelConsensusStateResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryChannelConsensusStateResponse copyWith(void Function(QueryChannelConsensusStateResponse) updates) => super.copyWith((message) => updates(message as QueryChannelConsensusStateResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryChannelConsensusStateResponse create() => QueryChannelConsensusStateResponse._();
  QueryChannelConsensusStateResponse createEmptyInstance() => create();
  static $pb.PbList<QueryChannelConsensusStateResponse> createRepeated() => $pb.PbList<QueryChannelConsensusStateResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryChannelConsensusStateResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryChannelConsensusStateResponse>(create);
  static QueryChannelConsensusStateResponse _defaultInstance;

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
  $core.String get clientId => $_getSZ(1);
  @$pb.TagNumber(2)
  set clientId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasClientId() => $_has(1);
  @$pb.TagNumber(2)
  void clearClientId() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get proof => $_getN(2);
  @$pb.TagNumber(3)
  set proof($core.List<$core.int> v) { $_setBytes(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasProof() => $_has(2);
  @$pb.TagNumber(3)
  void clearProof() => clearField(3);

  @$pb.TagNumber(4)
  $3.Height get proofHeight => $_getN(3);
  @$pb.TagNumber(4)
  set proofHeight($3.Height v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasProofHeight() => $_has(3);
  @$pb.TagNumber(4)
  void clearProofHeight() => clearField(4);
  @$pb.TagNumber(4)
  $3.Height ensureProofHeight() => $_ensure(3);
}

class QueryPacketCommitmentRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPacketCommitmentRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sequence', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryPacketCommitmentRequest._() : super();
  factory QueryPacketCommitmentRequest() => create();
  factory QueryPacketCommitmentRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPacketCommitmentRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPacketCommitmentRequest clone() => QueryPacketCommitmentRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPacketCommitmentRequest copyWith(void Function(QueryPacketCommitmentRequest) updates) => super.copyWith((message) => updates(message as QueryPacketCommitmentRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPacketCommitmentRequest create() => QueryPacketCommitmentRequest._();
  QueryPacketCommitmentRequest createEmptyInstance() => create();
  static $pb.PbList<QueryPacketCommitmentRequest> createRepeated() => $pb.PbList<QueryPacketCommitmentRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryPacketCommitmentRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPacketCommitmentRequest>(create);
  static QueryPacketCommitmentRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get sequence => $_getI64(2);
  @$pb.TagNumber(3)
  set sequence($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSequence() => $_has(2);
  @$pb.TagNumber(3)
  void clearSequence() => clearField(3);
}

class QueryPacketCommitmentResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPacketCommitmentResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'commitment', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proof', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryPacketCommitmentResponse._() : super();
  factory QueryPacketCommitmentResponse() => create();
  factory QueryPacketCommitmentResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPacketCommitmentResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPacketCommitmentResponse clone() => QueryPacketCommitmentResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPacketCommitmentResponse copyWith(void Function(QueryPacketCommitmentResponse) updates) => super.copyWith((message) => updates(message as QueryPacketCommitmentResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPacketCommitmentResponse create() => QueryPacketCommitmentResponse._();
  QueryPacketCommitmentResponse createEmptyInstance() => create();
  static $pb.PbList<QueryPacketCommitmentResponse> createRepeated() => $pb.PbList<QueryPacketCommitmentResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryPacketCommitmentResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPacketCommitmentResponse>(create);
  static QueryPacketCommitmentResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get commitment => $_getN(0);
  @$pb.TagNumber(1)
  set commitment($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCommitment() => $_has(0);
  @$pb.TagNumber(1)
  void clearCommitment() => clearField(1);

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

class QueryPacketCommitmentsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPacketCommitmentsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..aOM<$6.PageRequest>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $6.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryPacketCommitmentsRequest._() : super();
  factory QueryPacketCommitmentsRequest() => create();
  factory QueryPacketCommitmentsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPacketCommitmentsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPacketCommitmentsRequest clone() => QueryPacketCommitmentsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPacketCommitmentsRequest copyWith(void Function(QueryPacketCommitmentsRequest) updates) => super.copyWith((message) => updates(message as QueryPacketCommitmentsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPacketCommitmentsRequest create() => QueryPacketCommitmentsRequest._();
  QueryPacketCommitmentsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryPacketCommitmentsRequest> createRepeated() => $pb.PbList<QueryPacketCommitmentsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryPacketCommitmentsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPacketCommitmentsRequest>(create);
  static QueryPacketCommitmentsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $6.PageRequest get pagination => $_getN(2);
  @$pb.TagNumber(3)
  set pagination($6.PageRequest v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasPagination() => $_has(2);
  @$pb.TagNumber(3)
  void clearPagination() => clearField(3);
  @$pb.TagNumber(3)
  $6.PageRequest ensurePagination() => $_ensure(2);
}

class QueryPacketCommitmentsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPacketCommitmentsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..pc<$4.PacketState>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'commitments', $pb.PbFieldType.PM, subBuilder: $4.PacketState.create)
    ..aOM<$6.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $6.PageResponse.create)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryPacketCommitmentsResponse._() : super();
  factory QueryPacketCommitmentsResponse() => create();
  factory QueryPacketCommitmentsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPacketCommitmentsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPacketCommitmentsResponse clone() => QueryPacketCommitmentsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPacketCommitmentsResponse copyWith(void Function(QueryPacketCommitmentsResponse) updates) => super.copyWith((message) => updates(message as QueryPacketCommitmentsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPacketCommitmentsResponse create() => QueryPacketCommitmentsResponse._();
  QueryPacketCommitmentsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryPacketCommitmentsResponse> createRepeated() => $pb.PbList<QueryPacketCommitmentsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryPacketCommitmentsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPacketCommitmentsResponse>(create);
  static QueryPacketCommitmentsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.PacketState> get commitments => $_getList(0);

  @$pb.TagNumber(2)
  $6.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($6.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $6.PageResponse ensurePagination() => $_ensure(1);

  @$pb.TagNumber(3)
  $3.Height get height => $_getN(2);
  @$pb.TagNumber(3)
  set height($3.Height v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearHeight() => clearField(3);
  @$pb.TagNumber(3)
  $3.Height ensureHeight() => $_ensure(2);
}

class QueryPacketReceiptRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPacketReceiptRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sequence', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryPacketReceiptRequest._() : super();
  factory QueryPacketReceiptRequest() => create();
  factory QueryPacketReceiptRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPacketReceiptRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPacketReceiptRequest clone() => QueryPacketReceiptRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPacketReceiptRequest copyWith(void Function(QueryPacketReceiptRequest) updates) => super.copyWith((message) => updates(message as QueryPacketReceiptRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPacketReceiptRequest create() => QueryPacketReceiptRequest._();
  QueryPacketReceiptRequest createEmptyInstance() => create();
  static $pb.PbList<QueryPacketReceiptRequest> createRepeated() => $pb.PbList<QueryPacketReceiptRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryPacketReceiptRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPacketReceiptRequest>(create);
  static QueryPacketReceiptRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get sequence => $_getI64(2);
  @$pb.TagNumber(3)
  set sequence($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSequence() => $_has(2);
  @$pb.TagNumber(3)
  void clearSequence() => clearField(3);
}

class QueryPacketReceiptResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPacketReceiptResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOB(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'received')
    ..a<$core.List<$core.int>>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proof', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryPacketReceiptResponse._() : super();
  factory QueryPacketReceiptResponse() => create();
  factory QueryPacketReceiptResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPacketReceiptResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPacketReceiptResponse clone() => QueryPacketReceiptResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPacketReceiptResponse copyWith(void Function(QueryPacketReceiptResponse) updates) => super.copyWith((message) => updates(message as QueryPacketReceiptResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPacketReceiptResponse create() => QueryPacketReceiptResponse._();
  QueryPacketReceiptResponse createEmptyInstance() => create();
  static $pb.PbList<QueryPacketReceiptResponse> createRepeated() => $pb.PbList<QueryPacketReceiptResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryPacketReceiptResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPacketReceiptResponse>(create);
  static QueryPacketReceiptResponse _defaultInstance;

  @$pb.TagNumber(2)
  $core.bool get received => $_getBF(0);
  @$pb.TagNumber(2)
  set received($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(2)
  $core.bool hasReceived() => $_has(0);
  @$pb.TagNumber(2)
  void clearReceived() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get proof => $_getN(1);
  @$pb.TagNumber(3)
  set proof($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(3)
  $core.bool hasProof() => $_has(1);
  @$pb.TagNumber(3)
  void clearProof() => clearField(3);

  @$pb.TagNumber(4)
  $3.Height get proofHeight => $_getN(2);
  @$pb.TagNumber(4)
  set proofHeight($3.Height v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasProofHeight() => $_has(2);
  @$pb.TagNumber(4)
  void clearProofHeight() => clearField(4);
  @$pb.TagNumber(4)
  $3.Height ensureProofHeight() => $_ensure(2);
}

class QueryPacketAcknowledgementRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPacketAcknowledgementRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..a<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sequence', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryPacketAcknowledgementRequest._() : super();
  factory QueryPacketAcknowledgementRequest() => create();
  factory QueryPacketAcknowledgementRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPacketAcknowledgementRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPacketAcknowledgementRequest clone() => QueryPacketAcknowledgementRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPacketAcknowledgementRequest copyWith(void Function(QueryPacketAcknowledgementRequest) updates) => super.copyWith((message) => updates(message as QueryPacketAcknowledgementRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPacketAcknowledgementRequest create() => QueryPacketAcknowledgementRequest._();
  QueryPacketAcknowledgementRequest createEmptyInstance() => create();
  static $pb.PbList<QueryPacketAcknowledgementRequest> createRepeated() => $pb.PbList<QueryPacketAcknowledgementRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryPacketAcknowledgementRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPacketAcknowledgementRequest>(create);
  static QueryPacketAcknowledgementRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get sequence => $_getI64(2);
  @$pb.TagNumber(3)
  set sequence($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSequence() => $_has(2);
  @$pb.TagNumber(3)
  void clearSequence() => clearField(3);
}

class QueryPacketAcknowledgementResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPacketAcknowledgementResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'acknowledgement', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proof', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryPacketAcknowledgementResponse._() : super();
  factory QueryPacketAcknowledgementResponse() => create();
  factory QueryPacketAcknowledgementResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPacketAcknowledgementResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPacketAcknowledgementResponse clone() => QueryPacketAcknowledgementResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPacketAcknowledgementResponse copyWith(void Function(QueryPacketAcknowledgementResponse) updates) => super.copyWith((message) => updates(message as QueryPacketAcknowledgementResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPacketAcknowledgementResponse create() => QueryPacketAcknowledgementResponse._();
  QueryPacketAcknowledgementResponse createEmptyInstance() => create();
  static $pb.PbList<QueryPacketAcknowledgementResponse> createRepeated() => $pb.PbList<QueryPacketAcknowledgementResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryPacketAcknowledgementResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPacketAcknowledgementResponse>(create);
  static QueryPacketAcknowledgementResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get acknowledgement => $_getN(0);
  @$pb.TagNumber(1)
  set acknowledgement($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAcknowledgement() => $_has(0);
  @$pb.TagNumber(1)
  void clearAcknowledgement() => clearField(1);

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

class QueryPacketAcknowledgementsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPacketAcknowledgementsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..aOM<$6.PageRequest>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $6.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryPacketAcknowledgementsRequest._() : super();
  factory QueryPacketAcknowledgementsRequest() => create();
  factory QueryPacketAcknowledgementsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPacketAcknowledgementsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPacketAcknowledgementsRequest clone() => QueryPacketAcknowledgementsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPacketAcknowledgementsRequest copyWith(void Function(QueryPacketAcknowledgementsRequest) updates) => super.copyWith((message) => updates(message as QueryPacketAcknowledgementsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPacketAcknowledgementsRequest create() => QueryPacketAcknowledgementsRequest._();
  QueryPacketAcknowledgementsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryPacketAcknowledgementsRequest> createRepeated() => $pb.PbList<QueryPacketAcknowledgementsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryPacketAcknowledgementsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPacketAcknowledgementsRequest>(create);
  static QueryPacketAcknowledgementsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $6.PageRequest get pagination => $_getN(2);
  @$pb.TagNumber(3)
  set pagination($6.PageRequest v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasPagination() => $_has(2);
  @$pb.TagNumber(3)
  void clearPagination() => clearField(3);
  @$pb.TagNumber(3)
  $6.PageRequest ensurePagination() => $_ensure(2);
}

class QueryPacketAcknowledgementsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPacketAcknowledgementsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..pc<$4.PacketState>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'acknowledgements', $pb.PbFieldType.PM, subBuilder: $4.PacketState.create)
    ..aOM<$6.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $6.PageResponse.create)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryPacketAcknowledgementsResponse._() : super();
  factory QueryPacketAcknowledgementsResponse() => create();
  factory QueryPacketAcknowledgementsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPacketAcknowledgementsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPacketAcknowledgementsResponse clone() => QueryPacketAcknowledgementsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPacketAcknowledgementsResponse copyWith(void Function(QueryPacketAcknowledgementsResponse) updates) => super.copyWith((message) => updates(message as QueryPacketAcknowledgementsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPacketAcknowledgementsResponse create() => QueryPacketAcknowledgementsResponse._();
  QueryPacketAcknowledgementsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryPacketAcknowledgementsResponse> createRepeated() => $pb.PbList<QueryPacketAcknowledgementsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryPacketAcknowledgementsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPacketAcknowledgementsResponse>(create);
  static QueryPacketAcknowledgementsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$4.PacketState> get acknowledgements => $_getList(0);

  @$pb.TagNumber(2)
  $6.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($6.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $6.PageResponse ensurePagination() => $_ensure(1);

  @$pb.TagNumber(3)
  $3.Height get height => $_getN(2);
  @$pb.TagNumber(3)
  set height($3.Height v) { setField(3, v); }
  @$pb.TagNumber(3)
  $core.bool hasHeight() => $_has(2);
  @$pb.TagNumber(3)
  void clearHeight() => clearField(3);
  @$pb.TagNumber(3)
  $3.Height ensureHeight() => $_ensure(2);
}

class QueryUnreceivedPacketsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryUnreceivedPacketsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..p<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'packetCommitmentSequences', $pb.PbFieldType.PU6)
    ..hasRequiredFields = false
  ;

  QueryUnreceivedPacketsRequest._() : super();
  factory QueryUnreceivedPacketsRequest() => create();
  factory QueryUnreceivedPacketsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryUnreceivedPacketsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryUnreceivedPacketsRequest clone() => QueryUnreceivedPacketsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryUnreceivedPacketsRequest copyWith(void Function(QueryUnreceivedPacketsRequest) updates) => super.copyWith((message) => updates(message as QueryUnreceivedPacketsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryUnreceivedPacketsRequest create() => QueryUnreceivedPacketsRequest._();
  QueryUnreceivedPacketsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryUnreceivedPacketsRequest> createRepeated() => $pb.PbList<QueryUnreceivedPacketsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryUnreceivedPacketsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryUnreceivedPacketsRequest>(create);
  static QueryUnreceivedPacketsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$fixnum.Int64> get packetCommitmentSequences => $_getList(2);
}

class QueryUnreceivedPacketsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryUnreceivedPacketsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..p<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sequences', $pb.PbFieldType.PU6)
    ..aOM<$3.Height>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryUnreceivedPacketsResponse._() : super();
  factory QueryUnreceivedPacketsResponse() => create();
  factory QueryUnreceivedPacketsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryUnreceivedPacketsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryUnreceivedPacketsResponse clone() => QueryUnreceivedPacketsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryUnreceivedPacketsResponse copyWith(void Function(QueryUnreceivedPacketsResponse) updates) => super.copyWith((message) => updates(message as QueryUnreceivedPacketsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryUnreceivedPacketsResponse create() => QueryUnreceivedPacketsResponse._();
  QueryUnreceivedPacketsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryUnreceivedPacketsResponse> createRepeated() => $pb.PbList<QueryUnreceivedPacketsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryUnreceivedPacketsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryUnreceivedPacketsResponse>(create);
  static QueryUnreceivedPacketsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$fixnum.Int64> get sequences => $_getList(0);

  @$pb.TagNumber(2)
  $3.Height get height => $_getN(1);
  @$pb.TagNumber(2)
  set height($3.Height v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasHeight() => $_has(1);
  @$pb.TagNumber(2)
  void clearHeight() => clearField(2);
  @$pb.TagNumber(2)
  $3.Height ensureHeight() => $_ensure(1);
}

class QueryUnreceivedAcksRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryUnreceivedAcksRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..p<$fixnum.Int64>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'packetAckSequences', $pb.PbFieldType.PU6)
    ..hasRequiredFields = false
  ;

  QueryUnreceivedAcksRequest._() : super();
  factory QueryUnreceivedAcksRequest() => create();
  factory QueryUnreceivedAcksRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryUnreceivedAcksRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryUnreceivedAcksRequest clone() => QueryUnreceivedAcksRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryUnreceivedAcksRequest copyWith(void Function(QueryUnreceivedAcksRequest) updates) => super.copyWith((message) => updates(message as QueryUnreceivedAcksRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryUnreceivedAcksRequest create() => QueryUnreceivedAcksRequest._();
  QueryUnreceivedAcksRequest createEmptyInstance() => create();
  static $pb.PbList<QueryUnreceivedAcksRequest> createRepeated() => $pb.PbList<QueryUnreceivedAcksRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryUnreceivedAcksRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryUnreceivedAcksRequest>(create);
  static QueryUnreceivedAcksRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$fixnum.Int64> get packetAckSequences => $_getList(2);
}

class QueryUnreceivedAcksResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryUnreceivedAcksResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..p<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sequences', $pb.PbFieldType.PU6)
    ..aOM<$3.Height>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'height', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryUnreceivedAcksResponse._() : super();
  factory QueryUnreceivedAcksResponse() => create();
  factory QueryUnreceivedAcksResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryUnreceivedAcksResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryUnreceivedAcksResponse clone() => QueryUnreceivedAcksResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryUnreceivedAcksResponse copyWith(void Function(QueryUnreceivedAcksResponse) updates) => super.copyWith((message) => updates(message as QueryUnreceivedAcksResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryUnreceivedAcksResponse create() => QueryUnreceivedAcksResponse._();
  QueryUnreceivedAcksResponse createEmptyInstance() => create();
  static $pb.PbList<QueryUnreceivedAcksResponse> createRepeated() => $pb.PbList<QueryUnreceivedAcksResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryUnreceivedAcksResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryUnreceivedAcksResponse>(create);
  static QueryUnreceivedAcksResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$fixnum.Int64> get sequences => $_getList(0);

  @$pb.TagNumber(2)
  $3.Height get height => $_getN(1);
  @$pb.TagNumber(2)
  set height($3.Height v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasHeight() => $_has(1);
  @$pb.TagNumber(2)
  void clearHeight() => clearField(2);
  @$pb.TagNumber(2)
  $3.Height ensureHeight() => $_ensure(1);
}

class QueryNextSequenceReceiveRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryNextSequenceReceiveRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'portId')
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'channelId')
    ..hasRequiredFields = false
  ;

  QueryNextSequenceReceiveRequest._() : super();
  factory QueryNextSequenceReceiveRequest() => create();
  factory QueryNextSequenceReceiveRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryNextSequenceReceiveRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryNextSequenceReceiveRequest clone() => QueryNextSequenceReceiveRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryNextSequenceReceiveRequest copyWith(void Function(QueryNextSequenceReceiveRequest) updates) => super.copyWith((message) => updates(message as QueryNextSequenceReceiveRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryNextSequenceReceiveRequest create() => QueryNextSequenceReceiveRequest._();
  QueryNextSequenceReceiveRequest createEmptyInstance() => create();
  static $pb.PbList<QueryNextSequenceReceiveRequest> createRepeated() => $pb.PbList<QueryNextSequenceReceiveRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryNextSequenceReceiveRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryNextSequenceReceiveRequest>(create);
  static QueryNextSequenceReceiveRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get portId => $_getSZ(0);
  @$pb.TagNumber(1)
  set portId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPortId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPortId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get channelId => $_getSZ(1);
  @$pb.TagNumber(2)
  set channelId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasChannelId() => $_has(1);
  @$pb.TagNumber(2)
  void clearChannelId() => clearField(2);
}

class QueryNextSequenceReceiveResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryNextSequenceReceiveResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'ibc.core.channel.v1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'nextSequenceReceive', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$core.List<$core.int>>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proof', $pb.PbFieldType.OY)
    ..aOM<$3.Height>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'proofHeight', subBuilder: $3.Height.create)
    ..hasRequiredFields = false
  ;

  QueryNextSequenceReceiveResponse._() : super();
  factory QueryNextSequenceReceiveResponse() => create();
  factory QueryNextSequenceReceiveResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryNextSequenceReceiveResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryNextSequenceReceiveResponse clone() => QueryNextSequenceReceiveResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryNextSequenceReceiveResponse copyWith(void Function(QueryNextSequenceReceiveResponse) updates) => super.copyWith((message) => updates(message as QueryNextSequenceReceiveResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryNextSequenceReceiveResponse create() => QueryNextSequenceReceiveResponse._();
  QueryNextSequenceReceiveResponse createEmptyInstance() => create();
  static $pb.PbList<QueryNextSequenceReceiveResponse> createRepeated() => $pb.PbList<QueryNextSequenceReceiveResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryNextSequenceReceiveResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryNextSequenceReceiveResponse>(create);
  static QueryNextSequenceReceiveResponse _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get nextSequenceReceive => $_getI64(0);
  @$pb.TagNumber(1)
  set nextSequenceReceive($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasNextSequenceReceive() => $_has(0);
  @$pb.TagNumber(1)
  void clearNextSequenceReceive() => clearField(1);

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

