package types

// event types
//
//nolint:gosec // these aren't hard-coded credentials
const (
	AttributeAmount              = "amount"
	AttributeCreator             = "creator"
	AttributeSubdenom            = "subdenom"
	AttributeNewTokenDenom       = "new_token_denom"
	AttributeMintToAddress       = "mint_to_address"
	AttributeBurnFromAddress     = "burn_from_address"
	AttributeTransferFromAddress = "transfer_from_address"
	AttributeTransferToAddress   = "transfer_to_address"
	AttributeDenom               = "denom"
	AttributeNewAdmin            = "new_admin"
	AttributeDenomMetadata       = "denom_metadata"
)
