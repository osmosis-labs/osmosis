package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// EventTypeNeutronMessage defines the event type used by the Interchain Queries module events.
	EventTypeNeutronMessage = "neutron"

	// AttributeKeyQueryID represents the key for event attribute delivering the query ID of a
	// registered interchain query.
	AttributeKeyQueryID = "query_id"

	// AttributeKeyOwner represents the key for event attribute delivering the address of the
	// registrator of an interchain query.
	AttributeKeyOwner = "owner"

	// AttributeKeyConnectionID represents the key for event attribute delivering the connection ID
	// of an interchain query.
	AttributeKeyConnectionID = "connection_id"

	// AttributeKeyQueryType represents the key for event attribute delivering the query type
	// identifier (e.g. 'kv' or 'tx')
	AttributeKeyQueryType = "type"

	// AttributeKeyKVQuery represents the keys of the storage we want to get from remote chain for event attribute delivering the keys
	// of an interchain query.
	AttributeKeyKVQuery = "kv_key"

	// AttributeTransactionsFilterQuery represents the transactions filter for event attribute delivering the filter
	// of an interchain query.
	AttributeTransactionsFilterQuery = "tx_filter"

	// AttributeValueCategory represents the value for the 'module' event attribute.
	AttributeValueCategory = ModuleName

	// AttributeValueQueryUpdated represents the value for the 'action' event attribute.
	AttributeValueQueryUpdated = "query_updated"

	// AttributeValueQueryRemoved represents the value for the 'action' event attribute.
	AttributeValueQueryRemoved = "query_removed"

	// maxTransactionsFilters defines maximum allowed amount of tx filters in msgRegisterInterchainQuery
	maxTransactionsFilters = 32
)

const (
	InterchainQueryTypeKV InterchainQueryType = "kv"
	InterchainQueryTypeTX InterchainQueryType = "tx"

	kvPathKeyDelimiter = "/"
	kvKeysDelimiter    = ","
)

type InterchainQueryType string

func (icqt InterchainQueryType) IsValid() bool {
	return icqt.IsTX() || icqt.IsKV()
}

func (icqt InterchainQueryType) IsKV() bool {
	return icqt == InterchainQueryTypeKV
}

func (icqt InterchainQueryType) IsTX() bool {
	return icqt == InterchainQueryTypeTX
}

func (kv KVKey) ToString() string {
	return kv.Path + kvPathKeyDelimiter + hex.EncodeToString(kv.Key)
}

type KVKeys []*KVKey

func (keys KVKeys) String() string {
	if len(keys) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(keys[0].ToString())

	for _, key := range keys[1:] {
		b.WriteString(kvKeysDelimiter)
		b.WriteString(key.ToString())
	}

	return b.String()
}

// TransactionsFilter represents the model of transactions filter parameter used in interchain
// queries of type TX.
type TransactionsFilter []TransactionsFilterItem

// TransactionsFilterItem is a single condition for filtering transactions in search.
type TransactionsFilterItem struct {
	// Field is the field used in condition, e.g. tx.height or transfer.recipient.
	Field string `json:"field"`
	// Op is the operation for filtering, one of the following: eq, gt, gte, lt, lte.
	Op string `json:"op"`
	// Value is the value for comparison.
	Value interface{} `json:"value"`
}

// ValidateTransactionsFilter checks if the passed string is a valid TransactionsFilter value.
func ValidateTransactionsFilter(s string) error {
	const forbiddenCharacters = "\t\n\r\\()\"'=><"
	filters := TransactionsFilter{}
	if err := json.Unmarshal([]byte(s), &filters); err != nil {
		return fmt.Errorf("failed to unmarshal transactions filter: %w", err)
	}
	if len(filters) > maxTransactionsFilters {
		return fmt.Errorf("too many transactions filters, provided=%d, max=%d", len(filters), maxTransactionsFilters)
	}

	for idx, f := range filters {
		if strings.ContainsAny(f.Field, forbiddenCharacters) {
			return fmt.Errorf("transactions filter condition idx=%d is invalid: special symbols %s are not allowed", idx, forbiddenCharacters)
		}
		if f.Field == "" {
			return fmt.Errorf("transactions filter condition idx=%d is invalid: field couldn't be empty", idx)
		}
		switch value := f.Value.(type) {
		case string:
		case float64:
			// despite json turns numbers into float, decimals are not allowed by tendermint API
			if value != float64(int64(value)) {
				return fmt.Errorf("transactions filter condition idx=%d is invalid: value %v can't be a decimal number", idx, value)
			}
		default:
			return fmt.Errorf("transactions filter condition idx=%d is invalid: value '%v' is expected to be on of: string, number", idx, f.Value)
		}
		switch strings.ToLower(f.Op) {
		case "eq", "gt", "gte", "lt", "lte":
		default:
			return fmt.Errorf("transactions filter condition idx=%d is invalid: op '%s' is expected to be one of: eq, gt, gte, lt, lte", idx, f.Op)
		}
	}
	return nil
}
