# Upgrade History

The following table presents a list of the osmosis versions.

Each version is identified by a specific id, name, tag, block height and software upgrade proposal.

| ID    | Name      | Tag       | Starting Block | Release                                                                  | Proposal                                             |
|-------|-----------|-----------|----------------|--------------------------------------------------------------------------|------------------------------------------------------|
| `v3`  | Lithium   | `v3.1.0`  | 0              | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v3.1.0/)  | N.A. (Genesis)                                       |
| `v4`  | Berylium  | `v4.2.0`  | 1314500        | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v4.2.0/)  | [38](https://www.mintscan.io/osmosis/proposals/38)   |
| `v5`  | Boron     | `v6.4.0`  | 2383300        | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v6.4.0)   | [95](https://www.mintscan.io/osmosis/proposals/95)   |
| `v7`  | Carbon    | `v8.0.0`  | 3401000        | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v8.0.0/)  | [157](https://www.mintscan.io/osmosis/proposals/157) |
| `v9`  | Nitrogen  | `v10.1.1` | 4707300        | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v10.1.1/) | [252](https://www.mintscan.io/osmosis/proposals/252) |
| `v11` |           | `v11.0.1` | 5432450        | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v11.0.1/) | [296](https://www.mintscan.io/osmosis/proposals/296) |
| `v12` | Oxygen    | `v12.3.0` | 6246000        | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v12.3.0/) | [335](https://www.mintscan.io/osmosis/proposals/335) |
| `v13` | Fluorine  | `v13.1.2` | 7241500        | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v13.1.2/) | [370](https://www.mintscan.io/osmosis/proposals/370) |
| `v14` | Neon      | `v14.0.1` | 7937500        | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v14.0.1/) | [401](https://www.mintscan.io/osmosis/proposals/401) |
| `v15` | Sodium    | `v15.2.0` | 8732500        | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v15.2.0/) | [458](https://www.mintscan.io/osmosis/proposals/458) |
| `v16` | Magnesium | `v16.1.1` | 10517000       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v16.1.1/) | [556](https://www.mintscan.io/osmosis/proposals/556) |
| `v17` | Aluminium | `v17.0.0` | 11126100       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v17.0.0/) | [586](https://www.mintscan.io/osmosis/proposals/586) |
| `v18` |           | `v18.0.0` | 11155350       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v18.0.0/) | [588](https://www.mintscan.io/osmosis/proposals/588) |
| `v19` |           | `v19.2.0` | 11317300       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v19.2.0/) | [606](https://www.mintscan.io/osmosis/proposals/606) |
| `v20` | Silicon   | `v20.2.1` | 12028900       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v20.2.1/) | [658](https://www.mintscan.io/osmosis/proposals/658) |
| `v21` |           | `v21.1.4` | 12834100       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v21.1.4/) | [696](https://www.mintscan.io/osmosis/proposals/696) |
| `v22` |           | `v22.0.1` | 13325950       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v22.0.1/) | [714](https://www.mintscan.io/osmosis/proposals/714) |
| `v23` |           | `v23.0.0` | 13899375       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v23.0.0/) | [730](https://www.mintscan.io/osmosis/proposals/730) |
| `v24` |           | `v24.0.4` | 14830300       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v24.0.4/) | [763](https://www.mintscan.io/osmosis/proposals/763) |
| `v25` |           | `v25.2.1` | 15753500       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v25.2.1/) | [782](https://www.mintscan.io/osmosis/proposals/782) |
| `v26` |           | `v26.0.2` | 21046000       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v26.0.2/) | [837](https://www.mintscan.io/osmosis/proposals/837) |
| `v27` |           | `v27.0.1` | 24250100       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v27.0.1/) | [861](https://www.mintscan.io/osmosis/proposals/861) |
| `v28` |           | `v28.0.6` | 25861100       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v28.0.6/) | [879](https://www.mintscan.io/osmosis/proposals/879) |
| `v29` |           | `v29.0.2` | 33187000       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v29.0.2/) | [920](https://www.mintscan.io/osmosis/proposals/920) |
| `v30` |           | `v30.0.2` | 41332000       | [Release](https://github.com/osmosis-labs/osmosis/releases/tag/v30.0.2/) | [961](https://www.mintscan.io/osmosis/proposals/961) |

## Upgrade binaries

### v3.1.0

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v3.1.0/osmosisd-3.1.0-linux-amd64?checksum=sha256:6a73d75e9c75ea402c13edc8c5c4ed08e26c5d8e517d540a9ca8b7e7afa67f79",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v3.1.0/osmosisd-3.1.0-linux-arm64?checksum=sha256:893f8a9786ae76d4217260201cd94ab67010f68d98b9676a9b31c0a5e68d1eae"
  }
}
```

### v4.2.0

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v4.2.0/osmosisd-4.2.0-linux-amd64?checksum=sha256:a11c61a737983d176f23ce83fa5ff985000ce8d5107d738ee6fa7d59b8dd3053",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v4.2.0/osmosisd-4.2.0-linux-arm64?checksum=sha256:41260be15e874fbc6cc49757d9fe3d4e459634729e2b745923e508e9cb26f837"
  }
}
```

### v6.4.0

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v6.4.0/osmosisd-6.4.0-linux-amd64?checksum=sha256:e4017da5d1a0a3b37b4f6936ba7ef16f39972ae25f95feae43e506f14933cf94",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v6.4.0/osmosisd-6.4.0-linux-arm64?checksum=sha256:a101bb3feb0419293a3ecee17d732a312bf9e864a829905ed509c65b5944040b"
  }
}
```

### v8.0.0

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v8.0.0/osmosisd-8.0.0-linux-amd64?checksum=sha256:4559ffe7d1e83b1519c2d45a709d35a89b51f8b35f8bba3b58aef92e667e254c"
  }
}
```

### v10.1.1

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v10.1.1/osmosisd-10.1.1-linux-amd64?checksum=sha256:aeae58f8b0be86d5e6e3aec1a8774eab4947207c88c7d4f309c46da98f6694e8",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v10.1.1/osmosisd-10.1.1-linux-arm64?checksum=sha256:d2c672ffa9782687f91d8d03bd23fdf8bd2fbe8b79c9cfcf8e9d302a1238a12c"
  }
}
```

### v11.0.1

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v11.0.1/osmosisd-11.0.1-linux-amd64?checksum=sha256:41b8fd2345a5e5b77ee5ed9b9ec5370d94bd1b1aa0d4ac2ac0ab02ee98ddd0d8",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v11.0.1/osmosisd-11.0.1-linux-arm64?checksum=sha256:267776170495ecaa831238ea8994f8790a379663c9ae47a2e93e5beceafd8e1d"
  }
}
```

### v12.3.0

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v12.3.0/osmosisd-12.3.0-linux-amd64?checksum=sha256:958210c919d13c281896fa9773c323c5534f0fa46d74807154f737609a00db70",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v12.3.0/osmosisd-12.3.0-linux-arm64?checksum=sha256:a931618c8a839c30e5cecfd2a88055cda1d68cc68557fe3303fe14e2de3bef8f"
  }
}
```

### v13.1.2

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v13.1.2/osmosisd-13.1.2-linux-amd64?checksum=sha256:67ed53046667c72ec6bfe962bcb4d6b122610876b3adf75fb7820ce52c34872d",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v13.1.2/osmosisd-13.1.2-linux-arm64?checksum=sha256:ad35c2a8d55852fa28187a55bdeb983494c07923f2a8a9f4479fb044d8d62bd9"
  }
}
```

### v14.0.1

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v14.0.1/osmosisd-14.0.1-linux-amd64?checksum=sha256:2cc4172bcf000f0f06b30b16864d875a8de2ee12df994a593dfd52a506851bce",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v14.0.1/osmosisd-14.0.1-linux-arm64?checksum=sha256:9a44c17d239c8d9afd19d0ff0bd14ca883fb9e9fbf69aff18c2607ffa6bff378"
  }
}
```

### v15.2.0

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v15.2.0/osmosisd-15.2.0-linux-amd64?checksum=sha256:3aab2f2668cb5a713d5770e46a777ef01c433753378702d9ae941aa2d1ee5618",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v15.2.0/osmosisd-15.2.0-linux-arm64?checksum=sha256:e158d30707a0ea51482237f99676223e81ce5a353966a5c83791d2662a930f35"
  }
}
```

### v16.1.1

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v16.1.1/osmosisd-16.1.1-linux-amd64?checksum=sha256:0ec66e32584fff24b6d62fc9938c69ff1a1bbdd8641d2ec9e0fd084aaa767ed3",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v16.1.1/osmosisd-16.1.1-linux-arm64?checksum=sha256:e2ccc743dd66da91d1df1ae4ecf92b36d658575f4ff507d5056eb640804e0401",
  }
}
```

### v17.0.0

```json
{
  "binaries": {
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v17.0.0/osmosisd-17.0.0-linux-arm64?checksum=sha256:d5eeab6a15e2acd7e24e7caf4fe3336c35367ff376da6299d404defd09ce52f9",
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v17.0.0/osmosisd-17.0.0-linux-amd64?checksum=sha256:d7fe62ae33cf2f0b48a17eb8b02644dadd9924f15861ed622cd90cb1a038135b"
  }
}
```

### v18.0.0

```json
{
  "binaries": {
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v18.0.0/osmosisd-18.0.0-linux-arm64?checksum=sha256:6d02ac17c720c2b7e01d364a3303b8a04c81b9e52038e0f81e1806d0d254d96e",
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v18.0.0/osmosisd-18.0.0-linux-amd64?checksum=sha256:d83b4122e3ff9c428c8d6dcfe89718f5229f80e9976dbab2deefeb68dceb0f38"
  }
}
```

### v19.2.0

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v19.2.0/osmosisd-19.2.0-linux-amd64?checksum=sha256:723ff1c5349eb3c039c3dc5f55895bbde2e1499fe7c0a96960cc6fadeec814c4",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v19.2.0/osmosisd-19.2.0-linux-arm64?checksum=sha256:d933b893d537422164a25bf161d7f269a59ea26d37f398cdb7dd575a9ec33ed2"
  }
}
```

### v20.2.1

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v20.2.1/osmosisd-20.2.1-linux-amd64?checksum=sha256:4e60a870861ca17819fbcb49fff981b5731ec1121d7cbab43987c5f04ff099fa",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v20.2.1/osmosisd-20.2.1-linux-arm64?checksum=sha256:4e7fe2cc369a9eef28a8083414c2d7e0a8cb5eb5b75e913ded06ee457dff62bb"
  }
}
```

### v21.1.4

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v21.1.4/osmosisd-21.1.4-linux-amd64?checksum=sha256:518fd61873622d505640ab08edb788e307e6beb4f52476fab77661dd96860416",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v21.1.4/osmosisd-21.1.4-linux-arm64?checksum=sha256:cdbc163f4f045718e1464a82ada4d9d2511dc8c6c3fea11044cb8e675b6f86f7"
  }
}
```

### v22.0.1

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v22.0.1/osmosisd-22.0.1-linux-amd64?checksum=sha256:427588cbdd82752e6b31383493637029358f4550fcc71b81182334de2a54a20c",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v22.0.1/osmosisd-22.0.1-linux-arm64?checksum=sha256:3f50785becdd9e180cbe41b3eb97f8e6d16d0d4329c69a31cab5e0e1b5901c35"
  }
}
```

### v23.0.0

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v23.0.0/osmosisd-23.0.0-linux-amd64?checksum=sha256:db5e29c6565a0eca9692d0f138decda2ca7cdfb2943b3a2319cae691927ad595",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v23.0.0/osmosisd-23.0.0-linux-arm64?checksum=sha256:35d39fcf166b4a287bc32523ae60a6c8a708df974a0b7cc6e23a7612157fe466"
  }
}
```

### v24.0.4

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v24.0.4/osmosisd-24.0.4-linux-amd64?checksum=sha256:2e1b9f1485915025ce78bdcf6dbd47906de7b8d3ad64e47a3e87d2c8f137ba85",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v24.0.4/osmosisd-24.0.4-linux-arm64?checksum=sha256:2c1e03903ae652b218ca5a59bc17f8de1b8980263917f02ec06850a884185ebf"
  }
}
```

### v25.2.1

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v25.2.1/osmosisd-25.2.1-linux-amd64?checksum=sha256:b13533d9118d40fc612d1f708566eabef8d5b56918978ad26a410baf582d9974",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v25.2.1/osmosisd-25.2.1-linux-arm64?checksum=sha256:2217c4156e58b8c4e42b8a7042eb8598ecc5e72413c0c636807cbe857d935cca"
  }
}
```

### v26.0.2

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v26.0.2/osmosisd-26.0.2-linux-amd64?checksum=sha256:a72edc827551d55285421680651982aabed2ca9a2f732a55731531af5b15cf5b",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v26.0.2/osmosisd-26.0.2-linux-arm64?checksum=sha256:3293b649c599b9615fdf5d3f05a687150bdcbade31e20e26d5e5de1cd5dbbb94"
  }
}
```

### v27.0.1

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v27.0.1/osmosisd-27.0.1-linux-amd64?checksum=sha256:84b989b90bae4036f4eb3f9a8b00d808d9bb709a75926630a20d32c26b367a14",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v27.0.1/osmosisd-27.0.1-linux-arm64?checksum=sha256:f8144e7d9a67a08f460c86178df1198a0bdd111dd82889ccbf6d410b75b28430"
  }
}
```

### v28.0.6

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v28.0.6/osmosisd-28.0.6-linux-amd64?checksum=sha256:0fc943b34be983152ffb8f46543489d90111050be48f3eac2a5fcf4ee963ac45",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v28.0.6/osmosisd-28.0.6-linux-arm64?checksum=sha256:1dad4de3e53c563e34b997b1ccdd2373fec7b14d04a22fa1c79f94bb0ab02ea4"
  }
}
```

### v29.0.2

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v29.0.2/osmosisd-29.0.2-linux-amd64?checksum=sha256:9276c11c814c8b5731ef7b904c96530c6933e71b02e6eb11f99b4be2b9968c92",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v29.0.2/osmosisd-29.0.2-linux-arm64?checksum=sha256:e6a3c81ba5ba9da6598582d6c430618a4cb083c7552302412def141f846098d6"
  }
}
```

### v30.0.2

```json
{
  "binaries": {
    "linux/amd64": "https://github.com/osmosis-labs/osmosis/releases/download/v30.0.2/osmosisd-30.0.2-linux-amd64?checksum=sha256:e17d3635bf88c9859cbc2b3575006c2df38d93747402639ccc2f71424e3faffa",
    "linux/arm64": "https://github.com/osmosis-labs/osmosis/releases/download/v30.0.2/osmosisd-30.0.2-linux-arm64?checksum=sha256:bf853c52bf865080de4f3fb623214f2474bcbe24d879c93dd85d463e70bdc8f6"
  }
}
```

## Replay from Genesis using Cosmovisor

Assuming that your osmosis home it's already initialized with the desired genesis and configuration,
to replay the chain from genesis using Cosmovisor:

1. Install version `v1.2.0` from the official [repository](https://github.com/cosmos/cosmos-sdk/tree/main/tools/cosmovisor).

Alternatively, you can download the appropriate binary for your platform from our mirrors:

| Platform | Architecture | Cosmovisor Binary URL                                                                                      |
|----------|--------------|------------------------------------------------------------------------------------------------------------|
| linux    | amd64        | [Download](https://osmosis.fra1.digitaloceanspaces.com/binaries/cosmovisor/cosmovisor-v1.2.0-linux-amd64)  |
| linux    | arm64        | [Download](https://osmosis.fra1.digitaloceanspaces.com/binaries/cosmovisor/cosmovisor-v1.2.0-linux-arm64)  |

1. Initialize the Cosmovisor directory following the specific structure outlined below:

```bash
<COSMOVISOR_HOME>
   ├── genesis
   │   └── bin
   │       └── osmosisd
   └── upgrades
       ├── v11
       │   └── bin
       │       └── osmosisd
       ├── v12
       │   └── bin
       │       └── osmosisd
       ├── v13
       │   └── bin
       │       └── osmosisd
       ├── v14
       │   └── bin
       │       └── osmosisd
       ├── v15
       │   └── bin
       │       └── osmosisd
       ├── v16
       │   └── bin
       │       └── osmosisd
       ├── v17
       │   └── bin
       │       └── osmosisd
       ├── v18
       │   └── bin
       │       └── osmosisd
       ├── v19
       │   └── bin
       │       └── osmosisd
       ├── v20
       │   └── bin
       │       └── osmosisd
       ├── v21
       │   └── bin
       │       └── osmosisd
       ├── v22
       │   └── bin
       │       └── osmosisd
       ├── v23
       │   └── bin
       │       └── osmosisd
       ├── v24
       │   └── bin
       │       └── osmosisd
       ├── v25
       │   └── bin
       │       └── osmosisd
       ├── v26
       │   └── bin
       │       └── osmosisd
       ├── v27
       │   └── bin
       │       └── osmosisd
       ├── v28
       │   └── bin
       │       └── osmosisd
       ├── v29
       │   └── bin
       │       └── osmosisd
       ├── v30
       │   └── bin
       │       └── osmosisd
       ├── v4
       │   └── bin
       │       └── osmosisd
       ├── v5
       │   └── bin
       │       └── osmosisd
       ├── v7
       │   └── bin
       │       └── osmosisd
       └── v9
           └── bin
               └── osmosisd
```

You can utilize the provided script to download the required binaries and initialize the Cosmovisor directory:

```bash
# Define osmosis home
osmosis_home="$HOME/.osmosisd"

# List of versions and their URLs
versions_info=(
    "v3:https://github.com/osmosis-labs/osmosis/releases/download/v3.1.0/osmosisd-3.1.0-linux-amd64?checksum=sha256:6a73d75e9c75ea402c13edc8c5c4ed08e26c5d8e517d540a9ca8b7e7afa67f79"
    "v4:https://github.com/osmosis-labs/osmosis/releases/download/v4.2.0/osmosisd-4.2.0-linux-amd64?checksum=sha256:a11c61a737983d176f23ce83fa5ff985000ce8d5107d738ee6fa7d59b8dd3053"
    "v5:https://github.com/osmosis-labs/osmosis/releases/download/v6.4.0/osmosisd-6.4.0-linux-amd64?checksum=sha256:e4017da5d1a0a3b37b4f6936ba7ef16f39972ae25f95feae43e506f14933cf94"
    "v7:https://github.com/osmosis-labs/osmosis/releases/download/v8.0.0/osmosisd-8.0.0-linux-amd64?checksum=sha256:4559ffe7d1e83b1519c2d45a709d35a89b51f8b35f8bba3b58aef92e667e254c"
    "v9:https://github.com/osmosis-labs/osmosis/releases/download/v10.1.1/osmosisd-10.1.1-linux-amd64?checksum=sha256:aeae58f8b0be86d5e6e3aec1a8774eab4947207c88c7d4f309c46da98f6694e8"
    "v11:https://github.com/osmosis-labs/osmosis/releases/download/v11.0.1/osmosisd-11.0.1-linux-amd64?checksum=sha256:41b8fd2345a5e5b77ee5ed9b9ec5370d94bd1b1aa0d4ac2ac0ab02ee98ddd0d8"
    "v12:https://github.com/osmosis-labs/osmosis/releases/download/v12.3.0/osmosisd-12.3.0-linux-amd64?checksum=sha256:958210c919d13c281896fa9773c323c5534f0fa46d74807154f737609a00db70"
    "v13:https://github.com/osmosis-labs/osmosis/releases/download/v13.1.2/osmosisd-13.1.2-linux-amd64?checksum=sha256:67ed53046667c72ec6bfe962bcb4d6b122610876b3adf75fb7820ce52c34872d"
    "v14:https://github.com/osmosis-labs/osmosis/releases/download/v14.0.1/osmosisd-14.0.1-linux-amd64?checksum=sha256:2cc4172bcf000f0f06b30b16864d875a8de2ee12df994a593dfd52a506851bce"
    "v15:https://github.com/osmosis-labs/osmosis/releases/download/v15.2.0/osmosisd-15.2.0-linux-amd64?checksum=sha256:3aab2f2668cb5a713d5770e46a777ef01c433753378702d9ae941aa2d1ee5618"
    "v16:https://github.com/osmosis-labs/osmosis/releases/download/v16.1.1/osmosisd-16.1.1-linux-amd64?checksum=sha256:f838618633c1d42f593dc33d26b25842f5900961e987fc08570bb81a062e311d"
    "v17:https://github.com/osmosis-labs/osmosis/releases/download/v17.0.0/osmosisd-17.0.0-linux-amd64?checksum=sha256:d7fe62ae33cf2f0b48a17eb8b02644dadd9924f15861ed622cd90cb1a038135b"
    "v18:https://github.com/osmosis-labs/osmosis/releases/download/v18.0.0/osmosisd-18.0.0-linux-amd64?checksum=sha256:d83b4122e3ff9c428c8d6dcfe89718f5229f80e9976dbab2deefeb68dceb0f38"
    "v19:https://github.com/osmosis-labs/osmosis/releases/download/v19.2.0/osmosisd-19.2.0-linux-arm64?checksum=sha256:d933b893d537422164a25bf161d7f269a59ea26d37f398cdb7dd575a9ec33ed2"
    "v20:https://github.com/osmosis-labs/osmosis/releases/download/v20.2.1/osmosisd-20.2.1-linux-amd64?checksum=sha256:4e60a870861ca17819fbcb49fff981b5731ec1121d7cbab43987c5f04ff099fa"
    "v21:https://github.com/osmosis-labs/osmosis/releases/download/v21.1.4/osmosisd-21.1.4-linux-amd64?checksum=sha256:518fd61873622d505640ab08edb788e307e6beb4f52476fab77661dd96860416"
    "v22:https://github.com/osmosis-labs/osmosis/releases/download/v22.0.1/osmosisd-22.0.1-linux-amd64?checksum=sha256:427588cbdd82752e6b31383493637029358f4550fcc71b81182334de2a54a20c"
    "v23:https://github.com/osmosis-labs/osmosis/releases/download/v23.0.0/osmosisd-23.0.0-linux-amd64?checksum=sha256:db5e29c6565a0eca9692d0f138decda2ca7cdfb2943b3a2319cae691927ad595"
    "v24:https://github.com/osmosis-labs/osmosis/releases/download/v24.0.4/osmosisd-24.0.4-linux-amd64?checksum=sha256:2e1b9f1485915025ce78bdcf6dbd47906de7b8d3ad64e47a3e87d2c8f137ba85"
    "v25:https://github.com/osmosis-labs/osmosis/releases/download/v25.2.1/osmosisd-25.2.1-linux-amd64?checksum=sha256:b13533d9118d40fc612d1f708566eabef8d5b56918978ad26a410baf582d9974"
    "v26:https://github.com/osmosis-labs/osmosis/releases/download/v26.0.2/osmosisd-26.0.2-linux-amd64?checksum=sha256:a72edc827551d55285421680651982aabed2ca9a2f732a55731531af5b15cf5b"
    "v27:https://github.com/osmosis-labs/osmosis/releases/download/v27.0.1/osmosisd-27.0.1-linux-amd64?checksum=sha256:84b989b90bae4036f4eb3f9a8b00d808d9bb709a75926630a20d32c26b367a14"
    "v28:https://github.com/osmosis-labs/osmosis/releases/download/v28.0.6/osmosisd-28.0.6-linux-amd64?checksum=sha256:0fc943b34be983152ffb8f46543489d90111050be48f3eac2a5fcf4ee963ac45"
    "v29:https://github.com/osmosis-labs/osmosis/releases/download/v29.0.2/osmosisd-29.0.2-linux-amd64?checksum=sha256:9276c11c814c8b5731ef7b904c96530c6933e71b02e6eb11f99b4be2b9968c92"
    "v30:https://github.com/osmosis-labs/osmosis/releases/download/v30.0.2/osmosisd-30.0.2-linux-amd64?checksum=sha256:e17d3635bf88c9859cbc2b3575006c2df38d93747402639ccc2f71424e3faffa"
)

# Create the cosmovisor directory
echo "📁 Creating the cosmovisor directory: ${osmosis_home}/cosmovisor"
mkdir -p "${osmosis_home}/cosmovisor"

# Create the genesis directory and download v3 binary to /cosmovisor/genesis/bin
echo "📁 Creating the genesis directory: ${osmosis_home}/cosmovisor/genesis/bin"
mkdir -p "${osmosis_home}/cosmovisor/genesis/bin"

echo "⬇️ Downloading v3 binary to: ${osmosis_home}/cosmovisor/genesis/bin/osmosisd"
wget -q -O "${osmosis_home}/cosmovisor/genesis/bin/osmosisd" "$(echo ${versions_info[0]} | cut -d: -f2)"

# Create the upgrades directories for each version and download the binaries
for version_info in "${versions_info[@]:1}"; do
    version="${version_info%%:*}"
    binary_url="${version_info#*:}"
    echo
    echo "📁 Creating ${version} directory: ${osmosis_home}/cosmovisor/upgrades/${version}/bin"
    mkdir -p "${osmosis_home}/cosmovisor/upgrades/${version}/bin"
    echo "⬇️  Downloading ${version} binary to: ${osmosis_home}/cosmovisor/upgrades/${version}/bin/osmosisd"
    wget -q -O "${osmosis_home}/cosmovisor/upgrades/${version}/bin/osmosisd" "$binary_url"
done
```

3. Replaying the chain from historical data requires the presence of at least one archive nodes in the persistent peers. 
Ensure that you include the following configuration in your `config.toml` file:

```toml
[p2p]
persistent_peers = "37c195e518c001099f956202d34af029b04f2c97@65.109.20.216:26656" 
```

4. Run cosmovisor with `DAEMON_ALLOW_DOWNLOAD_BINARIES=false` 
