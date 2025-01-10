package cosmwasmpool

import (
	"github.com/Masterminds/semver"
)

// CosmWasm contract info from [cw2 spec](https://github.com/CosmWasm/cw-minus/blob/main/packages/cw2/README.md)
type ContractInfo struct {
	Contract string `json:"contract"`
	Version  string `json:"version"`
}

const (
	AlloyTranmuterName        = "crates.io:transmuter"
	AlloyTransmuterMinVersion = "3.0.0"
)

// Check if the contract info matches the given contract and version constrains
func (ci *ContractInfo) Matches(contract string, versionConstrains *semver.Constraints) bool {
	version, err := semver.NewVersion(ci.Version)
	validSemver := err == nil

	// matches only if:
	// - semver is valid
	// - contract matches
	// - version constrains matches
	return validSemver && (ci.Contract == contract && versionConstrains.Check(version))
}

func mustParseSemverConstraint(constraint string) *semver.Constraints {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}
	return c
}

// CosmWasmPoolModel is a model for the pool data of a CosmWasm pool
// It includes the contract info and the pool data
// The CWPoolData works like a tagged union to hold different types of data
// depending on the contract and its version
type CosmWasmPoolModel struct {
	ContractInfo ContractInfo     `json:"contract_info"`
	Data         CosmWasmPoolData `json:"data"`
}

// CosmWasmPoolData is the custom data for each type of CosmWasm pool
// This struct is intended to work like tagged union in other languages
// so that it can hold different types of data depending on the contract
type CosmWasmPoolData struct {
	// Data for AlloyTransmuter contract, must be present if and only if `IsAlloyTransmuter()` is true
	AlloyTransmuter *AlloyTransmuterData `json:"alloy_transmuter,omitempty"`

	// Data for Orderbook contract, must be present if and only if `IsOrderbook()` is true
	Orderbook *OrderbookData `json:"orderbook,omitempty"`
}

func NewCWPoolModel(contract string, version string, data CosmWasmPoolData) *CosmWasmPoolModel {
	return &CosmWasmPoolModel{
		ContractInfo: ContractInfo{
			Contract: contract,
			Version:  version,
		},
		Data: data,
	}
}
