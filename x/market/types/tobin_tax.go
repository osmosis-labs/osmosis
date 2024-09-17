package types

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"gopkg.in/yaml.v2"
)

// TobinTax - struct to store tobin tax for the specific denom with high volatility
type TobinTax struct {
	Denom   string       `json:"denom" yaml:"denom"`
	TaxRate osmomath.Dec `json:"tax_rate" yaml:"tax_rate"`
}

// String implements fmt.Stringer interface
func (tt TobinTax) String() string {
	out, _ := yaml.Marshal(tt)
	return string(out)
}

// TobinTaxList is convenience wrapper to handle TobinTax array
type TobinTaxList []TobinTax

// String implements fmt.Stringer interface
func (ttl TobinTaxList) String() string {
	out, _ := yaml.Marshal(ttl)
	return string(out)
}
