import 'package:alan/alan.dart';

class BaseEnv {
  late NetworkInfo _networkInfo;
  late String _baseApiUrl;
  late String _baseEthUrl;

  void setEnv({
    required String lcdUrl,
    required String grpcUrl,
    required String lcdPort,
    required String grpcPort,
    required String ethUrl,
  }) {
    _networkInfo = NetworkInfo(
      bech32Hrp: 'cosmos',
      lcdInfo: LCDInfo(host: lcdUrl, port: int.parse(lcdPort)),
      grpcInfo: GRPCInfo(host: grpcUrl, port: int.parse(grpcPort)),
    );
    _baseApiUrl = "$lcdUrl:$lcdPort";
    _baseEthUrl = ethUrl;
  }

  NetworkInfo get networkInfo => _networkInfo;

  String get baseApiUrl => _baseApiUrl;

  String get baseEthUrl => _baseEthUrl;
}

const lcdPort = String.fromEnvironment('LCD_PORT', defaultValue: '1317');
const grpcPort = String.fromEnvironment('GRPC_PORT', defaultValue: '9091');
const lcdUrl = String.fromEnvironment('LCD_URL', defaultValue: 'localhost');
const grpcUrl = String.fromEnvironment('GRPC_URL', defaultValue: 'localhost');
const ethUrl = String.fromEnvironment('ETH_URL', defaultValue: 'HTTP://127.0.0.1:7545');
