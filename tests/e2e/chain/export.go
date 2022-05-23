package chain

import "fmt"

type ChainMeta struct {
	DataDir string `json:"dataDir"`
	Id      string `json:"id"`
}

type Validator struct {
	Name          string `json:"name"`
	ConfigDir     string `json:"configDir"`
	Index         int    `json:"index"`
	Mnemonic      string `json:"mnemonic"`
	PublicAddress string `json:"publicAddress"`
	PublicKey     string `json:"publicKey"`
	OperAddress   string `json:"operAddress"`
}

type Chain struct {
	ChainMeta  ChainMeta    `json:"chainMeta"`
	Validators []*Validator `json:"validators"`
	PropNumber int          `json:"propNumber"`
}

func (c *ChainMeta) configDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, c.Id)
}
