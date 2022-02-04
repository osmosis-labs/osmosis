module github.com/osmosis-labs/osmosis

go 1.17

require (
	github.com/cosmos/cosmos-sdk v0.45.1
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/iavl v0.17.2
	github.com/cosmos/ibc-go/v2 v2.0.0
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/osmosis-labs/bech32-ibc v0.2.0-rc2
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.15
	github.com/tendermint/tm-db v0.6.6
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)

require github.com/dgraph-io/badger/v2 v2.2007.3 // indirect

replace (
	// Our cosmos-sdk branch is:  https://github.com/osmosis-labs/cosmos-sdk  v0.44.3x-osmo-v5
	github.com/cosmos/cosmos-sdk => github.com/osmosis-labs/cosmos-sdk v0.43.0-rc3.0.20220120015748-5df6adc097e8
	github.com/cosmos/ibc-go/v2 => github.com/osmosis-labs/ibc-go/v2 v2.0.2-osmo
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/tecbot/gorocksdb => github.com/cosmos/gorocksdb v1.1.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
