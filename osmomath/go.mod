module github.com/osmosis-labs/osmosis/osmomath

go 1.18

require (
	github.com/cosmos/cosmos-sdk v0.46.7
	github.com/stretchr/testify v1.8.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/armon/go-metrics v0.4.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd v0.22.1 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/confio/ics23/go v0.7.0 // indirect
	github.com/cosmos/btcutil v1.0.4 // indirect
	github.com/cosmos/gorocksdb v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/badger/v3 v3.2103.2 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/dustin/go-humanize v1.0.1-0.20200219035652-afde56e7acac // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-kit/kit v0.12.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/gogo/protobuf v1.3.3 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/flatbuffers v1.12.1 // indirect
	github.com/gtank/merlin v0.1.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20210601165009-122bf33a46e0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/cobra v1.6.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.14.0 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	github.com/tendermint/tendermint v0.34.24 // indirect
	github.com/tendermint/tm-db v0.6.8-0.20220506192307-f628bb5dc95b // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20221024183307-1bc688fe9f3e // indirect
	google.golang.org/grpc v1.50.1 // indirect
	google.golang.org/protobuf v1.28.2-0.20220831092852-f930b1dc76e8 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	// Our cosmos-sdk branch is:  https://github.com/osmosis-labs/cosmos-sdk, current branch: v13.x. Direct commit link: https://github.com/osmosis-labs/cosmos-sdk/commit/8757a61551aa1ea993c85a523e18094ab555b1d7
	// tag: https://github.com/osmosis-labs/cosmos-sdk/releases/tag/sdk-v13.0.0-rc2
	github.com/cosmos/cosmos-sdk => github.com/osmosis-labs/cosmos-sdk v0.45.1-0.20221118211718-545aed73e94e
	// use cosmos-compatible protobufs
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	// use grpc compatible with cosmos protobufs
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
