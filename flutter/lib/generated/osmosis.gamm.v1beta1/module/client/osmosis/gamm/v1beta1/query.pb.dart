///
//  Generated code. Do not modify.
//  source: osmosis/gamm/v1beta1/query.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../../google/protobuf/any.pb.dart' as $3;
import '../../../cosmos/base/query/v1beta1/pagination.pb.dart' as $8;
import 'pool.pb.dart' as $6;
import '../../../cosmos/base/v1beta1/coin.pb.dart' as $2;
import 'tx.pb.dart' as $0;

class QueryPoolRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPoolRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryPoolRequest._() : super();
  factory QueryPoolRequest() => create();
  factory QueryPoolRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPoolRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPoolRequest clone() => QueryPoolRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPoolRequest copyWith(void Function(QueryPoolRequest) updates) => super.copyWith((message) => updates(message as QueryPoolRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPoolRequest create() => QueryPoolRequest._();
  QueryPoolRequest createEmptyInstance() => create();
  static $pb.PbList<QueryPoolRequest> createRepeated() => $pb.PbList<QueryPoolRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryPoolRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPoolRequest>(create);
  static QueryPoolRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get poolId => $_getI64(0);
  @$pb.TagNumber(1)
  set poolId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPoolId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPoolId() => clearField(1);
}

class QueryPoolResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPoolResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOM<$3.Any>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pool', subBuilder: $3.Any.create)
    ..hasRequiredFields = false
  ;

  QueryPoolResponse._() : super();
  factory QueryPoolResponse() => create();
  factory QueryPoolResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPoolResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPoolResponse clone() => QueryPoolResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPoolResponse copyWith(void Function(QueryPoolResponse) updates) => super.copyWith((message) => updates(message as QueryPoolResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPoolResponse create() => QueryPoolResponse._();
  QueryPoolResponse createEmptyInstance() => create();
  static $pb.PbList<QueryPoolResponse> createRepeated() => $pb.PbList<QueryPoolResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryPoolResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPoolResponse>(create);
  static QueryPoolResponse _defaultInstance;

  @$pb.TagNumber(1)
  $3.Any get pool => $_getN(0);
  @$pb.TagNumber(1)
  set pool($3.Any v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasPool() => $_has(0);
  @$pb.TagNumber(1)
  void clearPool() => clearField(1);
  @$pb.TagNumber(1)
  $3.Any ensurePool() => $_ensure(0);
}

class QueryPoolsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPoolsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOM<$8.PageRequest>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $8.PageRequest.create)
    ..hasRequiredFields = false
  ;

  QueryPoolsRequest._() : super();
  factory QueryPoolsRequest() => create();
  factory QueryPoolsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPoolsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPoolsRequest clone() => QueryPoolsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPoolsRequest copyWith(void Function(QueryPoolsRequest) updates) => super.copyWith((message) => updates(message as QueryPoolsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPoolsRequest create() => QueryPoolsRequest._();
  QueryPoolsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryPoolsRequest> createRepeated() => $pb.PbList<QueryPoolsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryPoolsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPoolsRequest>(create);
  static QueryPoolsRequest _defaultInstance;

  @$pb.TagNumber(2)
  $8.PageRequest get pagination => $_getN(0);
  @$pb.TagNumber(2)
  set pagination($8.PageRequest v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(0);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $8.PageRequest ensurePagination() => $_ensure(0);
}

class QueryPoolsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPoolsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..pc<$3.Any>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pools', $pb.PbFieldType.PM, subBuilder: $3.Any.create)
    ..aOM<$8.PageResponse>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'pagination', subBuilder: $8.PageResponse.create)
    ..hasRequiredFields = false
  ;

  QueryPoolsResponse._() : super();
  factory QueryPoolsResponse() => create();
  factory QueryPoolsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPoolsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPoolsResponse clone() => QueryPoolsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPoolsResponse copyWith(void Function(QueryPoolsResponse) updates) => super.copyWith((message) => updates(message as QueryPoolsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPoolsResponse create() => QueryPoolsResponse._();
  QueryPoolsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryPoolsResponse> createRepeated() => $pb.PbList<QueryPoolsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryPoolsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPoolsResponse>(create);
  static QueryPoolsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$3.Any> get pools => $_getList(0);

  @$pb.TagNumber(2)
  $8.PageResponse get pagination => $_getN(1);
  @$pb.TagNumber(2)
  set pagination($8.PageResponse v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasPagination() => $_has(1);
  @$pb.TagNumber(2)
  void clearPagination() => clearField(2);
  @$pb.TagNumber(2)
  $8.PageResponse ensurePagination() => $_ensure(1);
}

class QueryNumPoolsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryNumPoolsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  QueryNumPoolsRequest._() : super();
  factory QueryNumPoolsRequest() => create();
  factory QueryNumPoolsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryNumPoolsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryNumPoolsRequest clone() => QueryNumPoolsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryNumPoolsRequest copyWith(void Function(QueryNumPoolsRequest) updates) => super.copyWith((message) => updates(message as QueryNumPoolsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryNumPoolsRequest create() => QueryNumPoolsRequest._();
  QueryNumPoolsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryNumPoolsRequest> createRepeated() => $pb.PbList<QueryNumPoolsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryNumPoolsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryNumPoolsRequest>(create);
  static QueryNumPoolsRequest _defaultInstance;
}

class QueryNumPoolsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryNumPoolsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'numPools', $pb.PbFieldType.OU6, protoName: 'numPools', defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryNumPoolsResponse._() : super();
  factory QueryNumPoolsResponse() => create();
  factory QueryNumPoolsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryNumPoolsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryNumPoolsResponse clone() => QueryNumPoolsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryNumPoolsResponse copyWith(void Function(QueryNumPoolsResponse) updates) => super.copyWith((message) => updates(message as QueryNumPoolsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryNumPoolsResponse create() => QueryNumPoolsResponse._();
  QueryNumPoolsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryNumPoolsResponse> createRepeated() => $pb.PbList<QueryNumPoolsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryNumPoolsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryNumPoolsResponse>(create);
  static QueryNumPoolsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get numPools => $_getI64(0);
  @$pb.TagNumber(1)
  set numPools($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasNumPools() => $_has(0);
  @$pb.TagNumber(1)
  void clearNumPools() => clearField(1);
}

class QueryPoolParamsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPoolParamsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryPoolParamsRequest._() : super();
  factory QueryPoolParamsRequest() => create();
  factory QueryPoolParamsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPoolParamsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPoolParamsRequest clone() => QueryPoolParamsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPoolParamsRequest copyWith(void Function(QueryPoolParamsRequest) updates) => super.copyWith((message) => updates(message as QueryPoolParamsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPoolParamsRequest create() => QueryPoolParamsRequest._();
  QueryPoolParamsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryPoolParamsRequest> createRepeated() => $pb.PbList<QueryPoolParamsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryPoolParamsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPoolParamsRequest>(create);
  static QueryPoolParamsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get poolId => $_getI64(0);
  @$pb.TagNumber(1)
  set poolId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPoolId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPoolId() => clearField(1);
}

class QueryPoolParamsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPoolParamsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOM<$6.PoolParams>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'params', subBuilder: $6.PoolParams.create)
    ..hasRequiredFields = false
  ;

  QueryPoolParamsResponse._() : super();
  factory QueryPoolParamsResponse() => create();
  factory QueryPoolParamsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPoolParamsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPoolParamsResponse clone() => QueryPoolParamsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPoolParamsResponse copyWith(void Function(QueryPoolParamsResponse) updates) => super.copyWith((message) => updates(message as QueryPoolParamsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPoolParamsResponse create() => QueryPoolParamsResponse._();
  QueryPoolParamsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryPoolParamsResponse> createRepeated() => $pb.PbList<QueryPoolParamsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryPoolParamsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPoolParamsResponse>(create);
  static QueryPoolParamsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $6.PoolParams get params => $_getN(0);
  @$pb.TagNumber(1)
  set params($6.PoolParams v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasParams() => $_has(0);
  @$pb.TagNumber(1)
  void clearParams() => clearField(1);
  @$pb.TagNumber(1)
  $6.PoolParams ensureParams() => $_ensure(0);
}

class QueryTotalSharesRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryTotalSharesRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryTotalSharesRequest._() : super();
  factory QueryTotalSharesRequest() => create();
  factory QueryTotalSharesRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryTotalSharesRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryTotalSharesRequest clone() => QueryTotalSharesRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryTotalSharesRequest copyWith(void Function(QueryTotalSharesRequest) updates) => super.copyWith((message) => updates(message as QueryTotalSharesRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryTotalSharesRequest create() => QueryTotalSharesRequest._();
  QueryTotalSharesRequest createEmptyInstance() => create();
  static $pb.PbList<QueryTotalSharesRequest> createRepeated() => $pb.PbList<QueryTotalSharesRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryTotalSharesRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryTotalSharesRequest>(create);
  static QueryTotalSharesRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get poolId => $_getI64(0);
  @$pb.TagNumber(1)
  set poolId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPoolId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPoolId() => clearField(1);
}

class QueryTotalSharesResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryTotalSharesResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOM<$2.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'totalShares', protoName: 'totalShares', subBuilder: $2.Coin.create)
    ..hasRequiredFields = false
  ;

  QueryTotalSharesResponse._() : super();
  factory QueryTotalSharesResponse() => create();
  factory QueryTotalSharesResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryTotalSharesResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryTotalSharesResponse clone() => QueryTotalSharesResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryTotalSharesResponse copyWith(void Function(QueryTotalSharesResponse) updates) => super.copyWith((message) => updates(message as QueryTotalSharesResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryTotalSharesResponse create() => QueryTotalSharesResponse._();
  QueryTotalSharesResponse createEmptyInstance() => create();
  static $pb.PbList<QueryTotalSharesResponse> createRepeated() => $pb.PbList<QueryTotalSharesResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryTotalSharesResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryTotalSharesResponse>(create);
  static QueryTotalSharesResponse _defaultInstance;

  @$pb.TagNumber(1)
  $2.Coin get totalShares => $_getN(0);
  @$pb.TagNumber(1)
  set totalShares($2.Coin v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasTotalShares() => $_has(0);
  @$pb.TagNumber(1)
  void clearTotalShares() => clearField(1);
  @$pb.TagNumber(1)
  $2.Coin ensureTotalShares() => $_ensure(0);
}

class QueryPoolAssetsRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPoolAssetsRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  QueryPoolAssetsRequest._() : super();
  factory QueryPoolAssetsRequest() => create();
  factory QueryPoolAssetsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPoolAssetsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPoolAssetsRequest clone() => QueryPoolAssetsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPoolAssetsRequest copyWith(void Function(QueryPoolAssetsRequest) updates) => super.copyWith((message) => updates(message as QueryPoolAssetsRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPoolAssetsRequest create() => QueryPoolAssetsRequest._();
  QueryPoolAssetsRequest createEmptyInstance() => create();
  static $pb.PbList<QueryPoolAssetsRequest> createRepeated() => $pb.PbList<QueryPoolAssetsRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryPoolAssetsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPoolAssetsRequest>(create);
  static QueryPoolAssetsRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get poolId => $_getI64(0);
  @$pb.TagNumber(1)
  set poolId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPoolId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPoolId() => clearField(1);
}

class QueryPoolAssetsResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryPoolAssetsResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..pc<$6.PoolAsset>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolAssets', $pb.PbFieldType.PM, protoName: 'poolAssets', subBuilder: $6.PoolAsset.create)
    ..hasRequiredFields = false
  ;

  QueryPoolAssetsResponse._() : super();
  factory QueryPoolAssetsResponse() => create();
  factory QueryPoolAssetsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryPoolAssetsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryPoolAssetsResponse clone() => QueryPoolAssetsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryPoolAssetsResponse copyWith(void Function(QueryPoolAssetsResponse) updates) => super.copyWith((message) => updates(message as QueryPoolAssetsResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryPoolAssetsResponse create() => QueryPoolAssetsResponse._();
  QueryPoolAssetsResponse createEmptyInstance() => create();
  static $pb.PbList<QueryPoolAssetsResponse> createRepeated() => $pb.PbList<QueryPoolAssetsResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryPoolAssetsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryPoolAssetsResponse>(create);
  static QueryPoolAssetsResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$6.PoolAsset> get poolAssets => $_getList(0);
}

class QuerySpotPriceRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QuerySpotPriceRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenInDenom', protoName: 'tokenInDenom')
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenOutDenom', protoName: 'tokenOutDenom')
    ..aOB(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'withSwapFee', protoName: 'withSwapFee')
    ..hasRequiredFields = false
  ;

  QuerySpotPriceRequest._() : super();
  factory QuerySpotPriceRequest() => create();
  factory QuerySpotPriceRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QuerySpotPriceRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QuerySpotPriceRequest clone() => QuerySpotPriceRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QuerySpotPriceRequest copyWith(void Function(QuerySpotPriceRequest) updates) => super.copyWith((message) => updates(message as QuerySpotPriceRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QuerySpotPriceRequest create() => QuerySpotPriceRequest._();
  QuerySpotPriceRequest createEmptyInstance() => create();
  static $pb.PbList<QuerySpotPriceRequest> createRepeated() => $pb.PbList<QuerySpotPriceRequest>();
  @$core.pragma('dart2js:noInline')
  static QuerySpotPriceRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QuerySpotPriceRequest>(create);
  static QuerySpotPriceRequest _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get poolId => $_getI64(0);
  @$pb.TagNumber(1)
  set poolId($fixnum.Int64 v) { $_setInt64(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPoolId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPoolId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get tokenInDenom => $_getSZ(1);
  @$pb.TagNumber(2)
  set tokenInDenom($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasTokenInDenom() => $_has(1);
  @$pb.TagNumber(2)
  void clearTokenInDenom() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get tokenOutDenom => $_getSZ(2);
  @$pb.TagNumber(3)
  set tokenOutDenom($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTokenOutDenom() => $_has(2);
  @$pb.TagNumber(3)
  void clearTokenOutDenom() => clearField(3);

  @$pb.TagNumber(4)
  $core.bool get withSwapFee => $_getBF(3);
  @$pb.TagNumber(4)
  set withSwapFee($core.bool v) { $_setBool(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasWithSwapFee() => $_has(3);
  @$pb.TagNumber(4)
  void clearWithSwapFee() => clearField(4);
}

class QuerySpotPriceResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QuerySpotPriceResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'spotPrice', protoName: 'spotPrice')
    ..hasRequiredFields = false
  ;

  QuerySpotPriceResponse._() : super();
  factory QuerySpotPriceResponse() => create();
  factory QuerySpotPriceResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QuerySpotPriceResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QuerySpotPriceResponse clone() => QuerySpotPriceResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QuerySpotPriceResponse copyWith(void Function(QuerySpotPriceResponse) updates) => super.copyWith((message) => updates(message as QuerySpotPriceResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QuerySpotPriceResponse create() => QuerySpotPriceResponse._();
  QuerySpotPriceResponse createEmptyInstance() => create();
  static $pb.PbList<QuerySpotPriceResponse> createRepeated() => $pb.PbList<QuerySpotPriceResponse>();
  @$core.pragma('dart2js:noInline')
  static QuerySpotPriceResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QuerySpotPriceResponse>(create);
  static QuerySpotPriceResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get spotPrice => $_getSZ(0);
  @$pb.TagNumber(1)
  set spotPrice($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSpotPrice() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpotPrice() => clearField(1);
}

class QuerySwapExactAmountInRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QuerySwapExactAmountInRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenIn', protoName: 'tokenIn')
    ..pc<$0.SwapAmountInRoute>(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'routes', $pb.PbFieldType.PM, subBuilder: $0.SwapAmountInRoute.create)
    ..hasRequiredFields = false
  ;

  QuerySwapExactAmountInRequest._() : super();
  factory QuerySwapExactAmountInRequest() => create();
  factory QuerySwapExactAmountInRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QuerySwapExactAmountInRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QuerySwapExactAmountInRequest clone() => QuerySwapExactAmountInRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QuerySwapExactAmountInRequest copyWith(void Function(QuerySwapExactAmountInRequest) updates) => super.copyWith((message) => updates(message as QuerySwapExactAmountInRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QuerySwapExactAmountInRequest create() => QuerySwapExactAmountInRequest._();
  QuerySwapExactAmountInRequest createEmptyInstance() => create();
  static $pb.PbList<QuerySwapExactAmountInRequest> createRepeated() => $pb.PbList<QuerySwapExactAmountInRequest>();
  @$core.pragma('dart2js:noInline')
  static QuerySwapExactAmountInRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QuerySwapExactAmountInRequest>(create);
  static QuerySwapExactAmountInRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get poolId => $_getI64(1);
  @$pb.TagNumber(2)
  set poolId($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPoolId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPoolId() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get tokenIn => $_getSZ(2);
  @$pb.TagNumber(3)
  set tokenIn($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasTokenIn() => $_has(2);
  @$pb.TagNumber(3)
  void clearTokenIn() => clearField(3);

  @$pb.TagNumber(4)
  $core.List<$0.SwapAmountInRoute> get routes => $_getList(3);
}

class QuerySwapExactAmountInResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QuerySwapExactAmountInResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenOutAmount', protoName: 'tokenOutAmount')
    ..hasRequiredFields = false
  ;

  QuerySwapExactAmountInResponse._() : super();
  factory QuerySwapExactAmountInResponse() => create();
  factory QuerySwapExactAmountInResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QuerySwapExactAmountInResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QuerySwapExactAmountInResponse clone() => QuerySwapExactAmountInResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QuerySwapExactAmountInResponse copyWith(void Function(QuerySwapExactAmountInResponse) updates) => super.copyWith((message) => updates(message as QuerySwapExactAmountInResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QuerySwapExactAmountInResponse create() => QuerySwapExactAmountInResponse._();
  QuerySwapExactAmountInResponse createEmptyInstance() => create();
  static $pb.PbList<QuerySwapExactAmountInResponse> createRepeated() => $pb.PbList<QuerySwapExactAmountInResponse>();
  @$core.pragma('dart2js:noInline')
  static QuerySwapExactAmountInResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QuerySwapExactAmountInResponse>(create);
  static QuerySwapExactAmountInResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get tokenOutAmount => $_getSZ(0);
  @$pb.TagNumber(1)
  set tokenOutAmount($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTokenOutAmount() => $_has(0);
  @$pb.TagNumber(1)
  void clearTokenOutAmount() => clearField(1);
}

class QuerySwapExactAmountOutRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QuerySwapExactAmountOutRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'sender')
    ..a<$fixnum.Int64>(2, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'poolId', $pb.PbFieldType.OU6, protoName: 'poolId', defaultOrMaker: $fixnum.Int64.ZERO)
    ..pc<$0.SwapAmountOutRoute>(3, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'routes', $pb.PbFieldType.PM, subBuilder: $0.SwapAmountOutRoute.create)
    ..aOS(4, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenOut', protoName: 'tokenOut')
    ..hasRequiredFields = false
  ;

  QuerySwapExactAmountOutRequest._() : super();
  factory QuerySwapExactAmountOutRequest() => create();
  factory QuerySwapExactAmountOutRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QuerySwapExactAmountOutRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QuerySwapExactAmountOutRequest clone() => QuerySwapExactAmountOutRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QuerySwapExactAmountOutRequest copyWith(void Function(QuerySwapExactAmountOutRequest) updates) => super.copyWith((message) => updates(message as QuerySwapExactAmountOutRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QuerySwapExactAmountOutRequest create() => QuerySwapExactAmountOutRequest._();
  QuerySwapExactAmountOutRequest createEmptyInstance() => create();
  static $pb.PbList<QuerySwapExactAmountOutRequest> createRepeated() => $pb.PbList<QuerySwapExactAmountOutRequest>();
  @$core.pragma('dart2js:noInline')
  static QuerySwapExactAmountOutRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QuerySwapExactAmountOutRequest>(create);
  static QuerySwapExactAmountOutRequest _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sender => $_getSZ(0);
  @$pb.TagNumber(1)
  set sender($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSender() => $_has(0);
  @$pb.TagNumber(1)
  void clearSender() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get poolId => $_getI64(1);
  @$pb.TagNumber(2)
  set poolId($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPoolId() => $_has(1);
  @$pb.TagNumber(2)
  void clearPoolId() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$0.SwapAmountOutRoute> get routes => $_getList(2);

  @$pb.TagNumber(4)
  $core.String get tokenOut => $_getSZ(3);
  @$pb.TagNumber(4)
  set tokenOut($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasTokenOut() => $_has(3);
  @$pb.TagNumber(4)
  void clearTokenOut() => clearField(4);
}

class QuerySwapExactAmountOutResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QuerySwapExactAmountOutResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..aOS(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'tokenInAmount', protoName: 'tokenInAmount')
    ..hasRequiredFields = false
  ;

  QuerySwapExactAmountOutResponse._() : super();
  factory QuerySwapExactAmountOutResponse() => create();
  factory QuerySwapExactAmountOutResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QuerySwapExactAmountOutResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QuerySwapExactAmountOutResponse clone() => QuerySwapExactAmountOutResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QuerySwapExactAmountOutResponse copyWith(void Function(QuerySwapExactAmountOutResponse) updates) => super.copyWith((message) => updates(message as QuerySwapExactAmountOutResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QuerySwapExactAmountOutResponse create() => QuerySwapExactAmountOutResponse._();
  QuerySwapExactAmountOutResponse createEmptyInstance() => create();
  static $pb.PbList<QuerySwapExactAmountOutResponse> createRepeated() => $pb.PbList<QuerySwapExactAmountOutResponse>();
  @$core.pragma('dart2js:noInline')
  static QuerySwapExactAmountOutResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QuerySwapExactAmountOutResponse>(create);
  static QuerySwapExactAmountOutResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get tokenInAmount => $_getSZ(0);
  @$pb.TagNumber(1)
  set tokenInAmount($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTokenInAmount() => $_has(0);
  @$pb.TagNumber(1)
  void clearTokenInAmount() => clearField(1);
}

class QueryTotalLiquidityRequest extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryTotalLiquidityRequest', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  QueryTotalLiquidityRequest._() : super();
  factory QueryTotalLiquidityRequest() => create();
  factory QueryTotalLiquidityRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryTotalLiquidityRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryTotalLiquidityRequest clone() => QueryTotalLiquidityRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryTotalLiquidityRequest copyWith(void Function(QueryTotalLiquidityRequest) updates) => super.copyWith((message) => updates(message as QueryTotalLiquidityRequest)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryTotalLiquidityRequest create() => QueryTotalLiquidityRequest._();
  QueryTotalLiquidityRequest createEmptyInstance() => create();
  static $pb.PbList<QueryTotalLiquidityRequest> createRepeated() => $pb.PbList<QueryTotalLiquidityRequest>();
  @$core.pragma('dart2js:noInline')
  static QueryTotalLiquidityRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryTotalLiquidityRequest>(create);
  static QueryTotalLiquidityRequest _defaultInstance;
}

class QueryTotalLiquidityResponse extends $pb.GeneratedMessage {
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'QueryTotalLiquidityResponse', package: const $pb.PackageName(const $core.bool.fromEnvironment('protobuf.omit_message_names') ? '' : 'osmosis.gamm.v1beta1'), createEmptyInstance: create)
    ..pc<$2.Coin>(1, const $core.bool.fromEnvironment('protobuf.omit_field_names') ? '' : 'liquidity', $pb.PbFieldType.PM, subBuilder: $2.Coin.create)
    ..hasRequiredFields = false
  ;

  QueryTotalLiquidityResponse._() : super();
  factory QueryTotalLiquidityResponse() => create();
  factory QueryTotalLiquidityResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory QueryTotalLiquidityResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  QueryTotalLiquidityResponse clone() => QueryTotalLiquidityResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  QueryTotalLiquidityResponse copyWith(void Function(QueryTotalLiquidityResponse) updates) => super.copyWith((message) => updates(message as QueryTotalLiquidityResponse)); // ignore: deprecated_member_use
  $pb.BuilderInfo get info_ => _i;
  @$core.pragma('dart2js:noInline')
  static QueryTotalLiquidityResponse create() => QueryTotalLiquidityResponse._();
  QueryTotalLiquidityResponse createEmptyInstance() => create();
  static $pb.PbList<QueryTotalLiquidityResponse> createRepeated() => $pb.PbList<QueryTotalLiquidityResponse>();
  @$core.pragma('dart2js:noInline')
  static QueryTotalLiquidityResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<QueryTotalLiquidityResponse>(create);
  static QueryTotalLiquidityResponse _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$2.Coin> get liquidity => $_getList(0);
}

