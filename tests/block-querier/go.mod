module blockquerier

go 1.21.1

replace (
	// force utilizing the following versions
	github.com/cosmos/cosmos-proto => github.com/cosmos/cosmos-proto v1.0.0-beta.2
	github.com/cosmos/cosmos-sdk => github.com/osmosis-labs/cosmos-sdk v0.47.6-0.20231108005754-ee4c51caf467
	github.com/cosmos/gogoproto => github.com/cosmos/gogoproto v1.4.10
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

	github.com/osmosis-labs/osmosis/v20 => ../../

	// replace as directed by sdk upgrading.md https://github.com/cosmos/cosmos-sdk/blob/393de266c8675dc16cc037c1a15011b1e990975f/UPGRADING.md?plain=1#L713
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7

	// newer versions of exp treat sorting differently, which is incompatible with the current version of cosmos-sdk
	golang.org/x/exp => golang.org/x/exp v0.0.0-20230711153332-06a737ee72cb
)

// exclusion so we use v1.0.0
exclude github.com/coinbase/rosetta-sdk-go v0.7.0
