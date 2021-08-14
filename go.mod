module github.com/osmosis-labs/osmosis

go 1.15

require (
	github.com/cosmos/cosmos-sdk v0.42.9
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/iavl v0.16.0
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.12
	github.com/tendermint/tm-db v0.6.4
	google.golang.org/genproto v0.0.0-20210114201628-6edceaf6022f
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.4.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4

replace github.com/tendermint/tendermint => github.com/tendermint/tendermint v0.34.12

replace github.com/cosmos/cosmos-sdk => github.com/osmosis-labs/cosmos-sdk v0.42.10-0.20210915013958-01114e89a579

replace github.com/tendermint/tm-db => github.com/osmosis-labs/tm-db v0.6.5-0.20210911033928-ba9154613417
