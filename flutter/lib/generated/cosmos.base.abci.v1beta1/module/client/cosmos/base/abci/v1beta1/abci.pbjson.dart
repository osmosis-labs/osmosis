///
//  Generated code. Do not modify.
//  source: cosmos/base/abci/v1beta1/abci.proto
//
// @dart = 2.3
// ignore_for_file: annotate_overrides,camel_case_types,unnecessary_const,non_constant_identifier_names,library_prefixes,unused_import,unused_shown_name,return_of_invalid_type,unnecessary_this,prefer_final_fields

const TxResponse$json = const {
  '1': 'TxResponse',
  '2': const [
    const {'1': 'height', '3': 1, '4': 1, '5': 3, '10': 'height'},
    const {'1': 'txhash', '3': 2, '4': 1, '5': 9, '8': const {}, '10': 'txhash'},
    const {'1': 'codespace', '3': 3, '4': 1, '5': 9, '10': 'codespace'},
    const {'1': 'code', '3': 4, '4': 1, '5': 13, '10': 'code'},
    const {'1': 'data', '3': 5, '4': 1, '5': 9, '10': 'data'},
    const {'1': 'raw_log', '3': 6, '4': 1, '5': 9, '10': 'rawLog'},
    const {'1': 'logs', '3': 7, '4': 3, '5': 11, '6': '.cosmos.base.abci.v1beta1.ABCIMessageLog', '8': const {}, '10': 'logs'},
    const {'1': 'info', '3': 8, '4': 1, '5': 9, '10': 'info'},
    const {'1': 'gas_wanted', '3': 9, '4': 1, '5': 3, '10': 'gasWanted'},
    const {'1': 'gas_used', '3': 10, '4': 1, '5': 3, '10': 'gasUsed'},
    const {'1': 'tx', '3': 11, '4': 1, '5': 11, '6': '.google.protobuf.Any', '10': 'tx'},
    const {'1': 'timestamp', '3': 12, '4': 1, '5': 9, '10': 'timestamp'},
  ],
  '7': const {},
};

const ABCIMessageLog$json = const {
  '1': 'ABCIMessageLog',
  '2': const [
    const {'1': 'msg_index', '3': 1, '4': 1, '5': 13, '10': 'msgIndex'},
    const {'1': 'log', '3': 2, '4': 1, '5': 9, '10': 'log'},
    const {'1': 'events', '3': 3, '4': 3, '5': 11, '6': '.cosmos.base.abci.v1beta1.StringEvent', '8': const {}, '10': 'events'},
  ],
  '7': const {},
};

const StringEvent$json = const {
  '1': 'StringEvent',
  '2': const [
    const {'1': 'type', '3': 1, '4': 1, '5': 9, '10': 'type'},
    const {'1': 'attributes', '3': 2, '4': 3, '5': 11, '6': '.cosmos.base.abci.v1beta1.Attribute', '8': const {}, '10': 'attributes'},
  ],
  '7': const {},
};

const Attribute$json = const {
  '1': 'Attribute',
  '2': const [
    const {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    const {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
};

const GasInfo$json = const {
  '1': 'GasInfo',
  '2': const [
    const {'1': 'gas_wanted', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'gasWanted'},
    const {'1': 'gas_used', '3': 2, '4': 1, '5': 4, '8': const {}, '10': 'gasUsed'},
  ],
};

const Result$json = const {
  '1': 'Result',
  '2': const [
    const {'1': 'data', '3': 1, '4': 1, '5': 12, '10': 'data'},
    const {'1': 'log', '3': 2, '4': 1, '5': 9, '10': 'log'},
    const {'1': 'events', '3': 3, '4': 3, '5': 11, '6': '.tendermint.abci.Event', '8': const {}, '10': 'events'},
  ],
  '7': const {},
};

const SimulationResponse$json = const {
  '1': 'SimulationResponse',
  '2': const [
    const {'1': 'gas_info', '3': 1, '4': 1, '5': 11, '6': '.cosmos.base.abci.v1beta1.GasInfo', '8': const {}, '10': 'gasInfo'},
    const {'1': 'result', '3': 2, '4': 1, '5': 11, '6': '.cosmos.base.abci.v1beta1.Result', '10': 'result'},
  ],
};

const MsgData$json = const {
  '1': 'MsgData',
  '2': const [
    const {'1': 'msg_type', '3': 1, '4': 1, '5': 9, '10': 'msgType'},
    const {'1': 'data', '3': 2, '4': 1, '5': 12, '10': 'data'},
  ],
  '7': const {},
};

const TxMsgData$json = const {
  '1': 'TxMsgData',
  '2': const [
    const {'1': 'data', '3': 1, '4': 3, '5': 11, '6': '.cosmos.base.abci.v1beta1.MsgData', '10': 'data'},
  ],
  '7': const {},
};

const SearchTxsResult$json = const {
  '1': 'SearchTxsResult',
  '2': const [
    const {'1': 'total_count', '3': 1, '4': 1, '5': 4, '8': const {}, '10': 'totalCount'},
    const {'1': 'count', '3': 2, '4': 1, '5': 4, '10': 'count'},
    const {'1': 'page_number', '3': 3, '4': 1, '5': 4, '8': const {}, '10': 'pageNumber'},
    const {'1': 'page_total', '3': 4, '4': 1, '5': 4, '8': const {}, '10': 'pageTotal'},
    const {'1': 'limit', '3': 5, '4': 1, '5': 4, '10': 'limit'},
    const {'1': 'txs', '3': 6, '4': 3, '5': 11, '6': '.cosmos.base.abci.v1beta1.TxResponse', '10': 'txs'},
  ],
  '7': const {},
};

