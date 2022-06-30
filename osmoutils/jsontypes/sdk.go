package jsontypes

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	d.Duration, err = time.ParseDuration(s)
	return err
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

type Coin struct {
	sdk.Coin
}

func (c *Coin) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	c.Coin, err = sdk.ParseCoinNormalized(s)
	return err
}

func (c Coin) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}
