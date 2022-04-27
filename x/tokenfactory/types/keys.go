package types

import (
	"strings"
)

const (
	// ModuleName defines the module name
	ModuleName = "tokenfactory"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_tokenfactory"
)

// KeySeparator is used to combine parts of the keys in the store
const KeySeparator = "|"

var (
	DenomAuthorityMetadataKey = "authoritymetadata"
	DenomsPrefixKey           = "denoms"
	CreatorPrefixKey          = "creator"
	AdminPrefixKey            = "admin"
)

func GetDenomPrefixStore(denom string) []byte {
	return []byte(strings.Join([]string{DenomsPrefixKey, denom, ""}, KeySeparator))
}

func GetCreatorPrefix(creator string) []byte {
	return []byte(strings.Join([]string{CreatorPrefixKey, creator, ""}, KeySeparator))
}

func GetCreatorsPrefix() []byte {
	return []byte(strings.Join([]string{CreatorPrefixKey, ""}, KeySeparator))
}
