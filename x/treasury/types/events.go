package types

// Treasury module event types
const (
	EventTypeTaxRateUpdate = "treasury_tax_rate_update"

	AttributeKeyOldTaxRate               = "old_tax_rate"
	AttributeKeyNewTaxRate               = "new_tax_rate"
	AttributeKeyExchangePoolRefillAmount = "exchange_pool_refill_amount"
	AttributeKeyRewardWeight             = "reward_weight"

	AttributeValueCategory = ModuleName
)
