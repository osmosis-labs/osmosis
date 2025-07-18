module github.com/osmosis-labs/osmosis/v30

go 1.23.2

require (
	cloud.google.com/go/pubsub v1.49.0
	cosmossdk.io/api v0.9.2
	cosmossdk.io/client/v2 v2.0.0-beta.6
	cosmossdk.io/core v0.12.1-0.20240725072823-6a2d039e1212
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/log v1.6.0
	cosmossdk.io/math v1.5.3
	cosmossdk.io/store v1.1.1
	cosmossdk.io/tools/confix v0.1.2
	cosmossdk.io/x/circuit v0.1.1
	cosmossdk.io/x/evidence v0.1.1
	cosmossdk.io/x/tx v0.13.7
	cosmossdk.io/x/upgrade v0.1.4
	github.com/CosmWasm/wasmd v0.53.3
	github.com/CosmWasm/wasmvm/v2 v2.2.4
	github.com/Masterminds/semver v1.5.0
	github.com/cometbft/cometbft v0.38.13
	github.com/cometbft/cometbft-db v0.14.1
	github.com/cosmos/cosmos-db v1.1.1
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/cosmos-sdk v0.50.14
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/gogoproto v1.7.0
	github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8 v8.2.0
	github.com/cosmos/ibc-apps/modules/async-icq/v8 v8.0.0
	github.com/cosmos/ibc-go/modules/capability v1.0.1
	github.com/cosmos/ibc-go/modules/light-clients/08-wasm v0.4.2-0.20240730185033-ccd4dc278e72
	github.com/cosmos/ibc-go/v8 v8.7.0
	github.com/cosmos/rosetta v0.50.12
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.4
	github.com/gorilla/mux v1.8.1
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-metrics v0.5.4
	github.com/iancoleman/orderedmap v0.3.0
	github.com/json-iterator/go v1.1.12
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/ory/dockertest/v3 v3.11.0
	github.com/osmosis-labs/go-mutesting v0.0.0-20221208041716-b43bcd97b3b3
	github.com/osmosis-labs/osmosis/osmomath v0.0.18
	github.com/osmosis-labs/osmosis/osmoutils v0.0.18
	github.com/osmosis-labs/osmosis/x/epochs v0.0.14
	github.com/osmosis-labs/osmosis/x/ibc-hooks v0.0.20
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/skip-mev/block-sdk/v2 v2.1.6
	github.com/spf13/cast v1.8.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.6
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.10.0
	github.com/tidwall/btree v1.7.0
	github.com/tidwall/gjson v1.18.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.59.0
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.36.0
	go.opentelemetry.io/otel/sdk v1.36.0
	go.uber.org/multierr v1.11.0
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56
	google.golang.org/genproto/googleapis/api v0.0.0-20250519155744-55703ea1f237
	google.golang.org/grpc v1.72.2
	google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools v2.2.0+incompatible
	mvdan.cc/gofumpt v0.8.0
)

require (
	cel.dev/expr v0.20.0 // indirect
	cloud.google.com/go v0.120.0 // indirect
	cloud.google.com/go/auth v0.15.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.7 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/iam v1.4.2 // indirect
	cloud.google.com/go/monitoring v1.24.0 // indirect
	cloud.google.com/go/storage v1.50.0 // indirect
	cosmossdk.io/collections v0.4.0 // indirect
	cosmossdk.io/depinject v1.1.0 // indirect
	cosmossdk.io/x/feegrant v0.1.1 // indirect
	dario.cat/mergo v1.0.0 // indirect
	github.com/DataDog/datadog-go v3.2.0+incompatible // indirect
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.26.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.50.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.50.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/aws/aws-sdk-go v1.44.327 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/bits-and-blooms/bitset v1.8.0 // indirect
	github.com/bytedance/sonic v1.13.1 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/cncf/xds/go v0.0.0-20250121191232-2f005788dc42 // indirect
	github.com/cockroachdb/apd/v2 v2.0.2 // indirect
	github.com/cockroachdb/errors v1.11.3 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240606204812-0bbfbd93a7ce // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.2 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/coinbase/rosetta-sdk-go/types v1.0.0 // indirect
	github.com/cosmos/gogogateway v1.2.0 // indirect
	github.com/cosmos/iavl v1.2.6 // indirect
	github.com/cosmos/ics23/go v0.11.0 // indirect
	github.com/cosmos/rosetta-sdk-go v0.10.0 // indirect
	github.com/creachadair/atomicfile v0.3.1 // indirect
	github.com/creachadair/tomledit v0.0.24 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/dgraph-io/badger/v4 v4.2.0 // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/emicklei/dot v1.6.2 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/getsentry/sentry-go v0.27.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/flatbuffers v23.5.26+incompatible // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/goware/urlx v0.3.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-getter v1.7.5 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-plugin v1.6.0 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/skiplist v1.2.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.8.14 // indirect
	github.com/manifoldco/promptui v0.9.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20230904125328-1f23a7beb09a // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/shamaton/msgpack/v2 v2.2.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spiffe/go-spiffe/v2 v2.5.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ulikunitz/xz v0.5.11 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	github.com/zimmski/go-mutesting v0.0.0-20210610104036-6d9217011a00 // indirect
	github.com/zondax/ledger-go v0.14.3 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.34.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.36.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.36.0 // indirect
	go.opentelemetry.io/proto/otlp v1.6.0 // indirect
	golang.org/x/arch v0.15.0 // indirect
	golang.org/x/oauth2 v0.28.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	google.golang.org/api v0.227.0 // indirect
	google.golang.org/genproto v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250519155744-55703ea1f237 // indirect
	gotest.tools/v3 v3.5.1 // indirect
	pgregory.net/rapid v1.1.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bgentry/speakeasy v0.1.1-0.20220910012023-760eaf8b6816 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/continuity v0.4.3 // indirect
	github.com/cosmos/btcutil v1.0.5
	github.com/cosmos/ledger-cosmos-go v0.14.0 // indirect
	github.com/danieljoos/wincred v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/docker/cli v26.1.4+incompatible // indirect
	github.com/docker/docker v27.1.1+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-kit/kit v0.13.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.2.4 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/orderedcode v0.0.1 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/handlers v1.5.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/hdevalence/ed25519consensus v0.1.0 // indirect
	github.com/improbable-eng/grpc-web v0.15.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/joho/godotenv v1.5.1
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/highwayhash v1.0.3 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/opencontainers/runc v1.1.14 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.20.5
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d
	github.com/tendermint/go-amino v0.16.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zimmski/go-tool v0.0.0-20150119110811-2dfdc9ac8439 // indirect
	github.com/zimmski/osutil v0.0.0-20190128123334-0d0b3ca231ac // indirect
	github.com/zondax/hid v0.9.2 // indirect
	go.etcd.io/bbolt v1.4.0-alpha.0.0.20240404170359-43604f3112c5 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.32.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	nhooyr.io/websocket v1.8.10 // indirect
)

replace (
	// TODO(https://github.com/cosmos/rosetta/issues/76): Rosetta requires cosmossdk.io/core v0.12.0 erroneously but
	// should use v0.11.0. The Cosmos build fails with types/context.go:65:29: undefined: comet.BlockInfo otherwise.
	cosmossdk.io/core => cosmossdk.io/core v0.11.0

	// Needs to be replaced due to iavlFastNodeModuleWhitelist feature
	// Disabling fast nodes makes nodes sync faster.
	// All nodes need to have the lockup fast nodes enabled though or else we process epoch slowly.
	// Also, snapshot nodes need to have all fast nodes enabled in order to prune quickly.
	// Also, we need an osmosis version of the store to enable aysnc pruning.
	// Direct cosmos-sdk branch link: https://github.com/osmosis-labs/cosmos-sdk/tree/osmo-v28/0.50.11/store, current branch: osmo-v28/0.50.11
	// Direct commit link: https://github.com/osmosis-labs/cosmos-sdk/commit/eb1a8e88a4ddf77bc2fe235fc07c57016b7386f0
	// Direct tag link: https://github.com/osmosis-labs/cosmos-sdk/releases/tag/store/v1.1.1-v0.50.11-v28-osmo-2
	cosmossdk.io/store => github.com/osmosis-labs/cosmos-sdk/store v1.1.1-v0.50.11-v28-osmo-2

	// coinbase/rosetta-sdk-go is not longer maintained/available, we use the mesh-sdk-go fork instead
	github.com/coinbase/rosetta-sdk-go/types v1.0.0 => github.com/coinbase/mesh-sdk-go/types v1.0.0-beta1

	// Direct cometbft branch link: https://github.com/osmosis-labs/cometbft/tree/osmo-v28/0.38.17, current branch: osmo-v28/v0.38.17
	// Direct commit link: https://github.com/osmosis-labs/cometbft/commit/58bfcdcb2fdd81655586678c2062cdc51adbdf0d
	// Direct tag link: https://github.com/osmosis-labs/cometbft/releases/tag/v0.38.17-v28-osmo-1
	github.com/cometbft/cometbft => github.com/osmosis-labs/cometbft v0.38.17-v28-osmo-1

	// Direct cosmos-sdk branch link: https://github.com/osmosis-labs/cosmos-sdk/tree/osmo-v30/0.50.14, current branch: osmo-v30/0.50.14
	// Direct commit link: https://github.com/osmosis-labs/cosmos-sdk/commit/1f78f02de9b2f60779c5062686201b45361fcb3f
	// Direct tag link: https://github.com/osmosis-labs/cosmos-sdk/releases/tag/v0.50.14-v30-osmo
	github.com/cosmos/cosmos-sdk => github.com/osmosis-labs/cosmos-sdk v0.50.14-v30-osmo

	// The block-sdk is no longer maintained by Skip, instead we use the osmosis-labs fork.
	// Direct block-sdk branch link: https://github.com/osmosis-labs/block-sdk/tree/release/v2.x.x, current branch: release/v2.x.x
	// Direct commit link: https://github.com/osmosis-labs/block-sdk/commit/a86e729f843d36392393a8ff6a14c9fbc15f2fcc
	// Direct tag link: https://github.com/osmosis-labs/block-sdk/releases/tag/v2.1.6
	github.com/skip-mev/block-sdk/v2 => github.com/osmosis-labs/block-sdk/v2 v2.1.6

	// replace as directed by sdk upgrading.md https://github.com/cosmos/cosmos-sdk/blob/393de266c8675dc16cc037c1a15011b1e990975f/UPGRADING.md?plain=1#L713
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
// // Local replaces commented for development
// github.com/osmosis-labs/osmosis/osmomath => ./osmomath
// github.com/osmosis-labs/osmosis/osmoutils => ./osmoutils
// github.com/osmosis-labs/osmosis/x/epochs => ./x/epochs
// github.com/osmosis-labs/osmosis/x/ibc-hooks => ./x/ibc-hooks
)

// exclusion so we use v1.0.0
exclude github.com/coinbase/rosetta-sdk-go v0.7.9

exclude github.com/gogo/protobuf v1.3.3
