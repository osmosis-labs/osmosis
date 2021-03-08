package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/tendermint/tendermint/crypto"
	"gopkg.in/yaml.v2"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// PoolAccountI defines an account interface for pools that hold tokens.
type PoolAccountI interface {
	authtypes.AccountI

	GetId() uint64
	GetPoolParams() PoolParams
	GetTotalWeight() sdk.Int
	GetTotalShare() sdk.Coin
	AddTotalShare(amt sdk.Int)
	SubTotalShare(amt sdk.Int)
	AddRecords(records []Record) error
	GetRecord(denom string) (Record, error)
	SetRecord(denom string, record Record) error
	GetRecords(denoms ...string) ([]Record, error)
	SetRecords(record []Record) error
	GetAllRecords() []Record
	SetTokenWeight(denom string, weight sdk.Int) error
	GetTokenWeight(denom string) (sdk.Int, error)
	SetTokenBalance(denom string, amount sdk.Int) error
	GetTokenBalance(denom string) (sdk.Int, error)
	LenRecords() int
}

var (
	// TODO: Add `GenesisAccount` type
	_ PoolAccountI = (*PoolAccount)(nil)
)

func NewPoolAddress(poolId uint64) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash(append(PoolAddressPrefix, sdk.Uint64ToBigEndian(poolId)...)))
}

func NewPoolAccount(poolId uint64, poolParams PoolParams) PoolAccountI {
	poolAddr := NewPoolAddress(poolId)
	baseAcc := authtypes.NewBaseAccountWithAddress(poolAddr)

	err := poolParams.Validate()
	if err != nil {
		panic(err)
	}

	return &PoolAccount{
		BaseAccount: baseAcc,
		Id:          poolId,
		PoolParams:  poolParams,
		TotalWeight: sdk.ZeroInt(),
		TotalShare:  sdk.NewCoin(fmt.Sprintf("osmosis/pool/%d", poolId), sdk.ZeroInt()),
		Records:     nil,
	}
}

func (params PoolParams) Validate() error {
	if params.ExitFee.LT(sdk.NewDec(0)) {
		return ErrNegativeExitFee
	}

	if params.ExitFee.GTE(sdk.NewDec(1)) {
		return ErrTooMuchExitFee
	}

	if params.SwapFee.LT(sdk.NewDec(0)) {
		return ErrNegativeSwapFee
	}

	if params.SwapFee.GTE(sdk.NewDec(1)) {
		return ErrTooMuchSwapFee
	}

	return nil
}

func (pa PoolAccount) GetId() uint64 {
	return pa.Id
}

func (pa PoolAccount) GetPoolParams() PoolParams {
	return pa.PoolParams
}

func (pa PoolAccount) GetTotalWeight() sdk.Int {
	return pa.TotalWeight
}

func (pa PoolAccount) GetTotalShare() sdk.Coin {
	return pa.TotalShare
}

func (pa *PoolAccount) AddTotalShare(amt sdk.Int) {
	pa.TotalShare.Amount = pa.TotalShare.Amount.Add(amt)
}

func (pa *PoolAccount) SubTotalShare(amt sdk.Int) {
	pa.TotalShare.Amount = pa.TotalShare.Amount.Sub(amt)
}

// AddRecords adds the records to the pool. If the same denom's record exists, will return error.
// And, records have to be sorted to search the denom's record by the binary search.
func (pa *PoolAccount) AddRecords(records []Record) error {
	exists := make(map[string]bool)
	for _, record := range pa.Records {
		exists[record.Token.Denom] = true
	}

	addTotalWeight := sdk.ZeroInt()

	for _, record := range records {
		if record.Token.Amount.LTE(sdk.ZeroInt()) {
			return fmt.Errorf("can't add the zero or negative balance of token")
		}

		if record.Weight.LTE(sdk.ZeroInt()) {
			return fmt.Errorf("can't add the zero or negative weight of token")
		}

		if exists[record.Token.Denom] {
			return fmt.Errorf("same record already exists")
		}
		exists[record.Token.Denom] = true

		addTotalWeight = addTotalWeight.Add(record.Weight)
	}

	pa.Records = append(pa.Records, records...)
	sort.Slice(pa.Records, func(i, j int) bool {
		recordA := pa.Records[i]
		recordB := pa.Records[j]

		return strings.Compare(recordA.Token.Denom, recordB.Token.Denom) == -1
	})

	pa.TotalWeight = pa.TotalWeight.Add(addTotalWeight)

	return nil
}

// GetRecords returns the denom's record, If the record doesn't exist, will return error.
// As above, it will search the denom's record by using binary search.
// So, it is important to make sure that the records are sorted.
func (pa PoolAccount) GetRecord(denom string) (Record, error) {
	if denom == "" {
		return Record{}, fmt.Errorf("you tried to find the record with empty denom")
	}

	if len(pa.Records) == 0 {
		return Record{}, fmt.Errorf("can't find the record (%s)", denom)
	}

	i := sort.Search(len(pa.Records), func(i int) bool {
		recordA := pa.Records[i]

		compare := strings.Compare(recordA.Token.Denom, denom)
		return compare == 1 || compare == 0
	})

	if i < 0 || i >= len(pa.Records) {
		return Record{}, fmt.Errorf("can't find the record (%s)", denom)
	}

	if pa.Records[i].Token.Denom != denom {
		return Record{}, fmt.Errorf("can't find the record (%s)", denom)
	}

	return pa.Records[i], nil
}

func (pa *PoolAccount) SetRecord(denom string, record Record) error {
	// Check that record exists.
	_, err := pa.GetRecord(denom)
	if err != nil {
		return err
	}

	if record.Token.Amount.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("can't add the zero or negative balance of token")
	}

	if record.Weight.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("can't add the zero or negative weight of token")
	}

	for i, oldRecord := range pa.Records {
		if oldRecord.Token.Denom == record.Token.Denom {
			deltaTokenWeight := record.Weight.Sub(oldRecord.Weight)

			pa.TotalWeight = pa.TotalWeight.Add(deltaTokenWeight)

			pa.Records[i] = record

			return nil
		}
	}

	return fmt.Errorf("can't find the record (%s)", denom)
}

func (pa *PoolAccount) SetRecords(records []Record) error {
	exists := make(map[string]int)
	for index, record := range pa.Records {
		exists[record.Token.Denom] = index
	}

	addingRecordsExists := make(map[string]bool)

	deltaTotalWeight := sdk.ZeroInt()

	for _, record := range records {
		if record.Token.Amount.LTE(sdk.ZeroInt()) {
			return fmt.Errorf("can't set the zero or negative balance of token")
		}

		if record.Weight.LTE(sdk.ZeroInt()) {
			return fmt.Errorf("can't set the zero or negative weight of token")
		}

		index, ok := exists[record.Token.Denom]
		if !ok {
			return fmt.Errorf("record doesn't exists")
		}

		if addingRecordsExists[record.Token.Denom] {
			return fmt.Errorf("adding records duplicated")
		}
		addingRecordsExists[record.Token.Denom] = true

		oldRecord := pa.Records[index]
		deltaTotalWeight = deltaTotalWeight.Add(record.Weight.Sub(oldRecord.Weight))

		pa.Records[index].Weight = record.Weight
		pa.Records[index].Token = record.Token
	}

	pa.TotalWeight = pa.TotalWeight.Add(deltaTotalWeight)

	return nil
}

func (pa PoolAccount) GetRecords(denoms ...string) ([]Record, error) {
	result := make([]Record, 0, len(denoms))

	for _, denom := range denoms {
		record, err := pa.GetRecord(denom)
		if err != nil {
			return nil, err
		}

		result = append(result, record)
	}

	return result, nil
}

func (pa PoolAccount) GetAllRecords() []Record {
	copyslice := make([]Record, len(pa.Records))
	copy(copyslice, pa.Records)
	return copyslice
}

func (pa *PoolAccount) SetTokenWeight(denom string, weight sdk.Int) error {
	record, err := pa.GetRecord(denom)
	if err != nil {
		return err
	}

	record.Weight = weight

	return pa.SetRecord(denom, record)
}

func (pa PoolAccount) GetTokenWeight(denom string) (sdk.Int, error) {
	record, err := pa.GetRecord(denom)
	if err != nil {
		return sdk.Int{}, err
	}

	return record.Weight, nil
}

func (pa *PoolAccount) SetTokenBalance(denom string, amount sdk.Int) error {
	record, err := pa.GetRecord(denom)
	if err != nil {
		return err
	}

	record.Token.Amount = amount

	return pa.SetRecord(denom, record)
}

func (pa PoolAccount) GetTokenBalance(denom string) (sdk.Int, error) {
	record, err := pa.GetRecord(denom)
	if err != nil {
		return sdk.Int{}, err
	}

	return record.Token.Amount, nil
}

func (pa PoolAccount) LenRecords() int {
	return len(pa.Records)
}

// SetPubKey - Implements AccountI
func (pa PoolAccount) SetPubKey(pubKey cryptotypes.PubKey) error {
	return fmt.Errorf("not supported for pool accounts")
}

// SetSequence - Implements AccountI
func (pa PoolAccount) SetSequence(seq uint64) error {
	return fmt.Errorf("not supported for pool accounts")
}

type poolAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
	Id            uint64         `json:"id" yaml:"id"`
	PoolParams    PoolParams     `json:"pool_params" yaml:"pool_params"`
	TotalWeight   sdk.Int        `json:"total_weight" yaml:"total_weight"`
	TotalShare    sdk.Coin       `json:"total_share" yaml:"total_share"`
	Records       []Record       `json:"records" yaml:"records"`
}

func (pa PoolAccount) String() string {
	out, _ := pa.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a PoolAccount.
func (pa PoolAccount) MarshalYAML() (interface{}, error) {
	accAddr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		return nil, err
	}

	bs, err := yaml.Marshal(poolAccountPretty{
		Address:       accAddr,
		PubKey:        "",
		AccountNumber: pa.AccountNumber,
		Id:            pa.Id,
		PoolParams:    pa.PoolParams,
		TotalWeight:   pa.TotalWeight,
		TotalShare:    pa.TotalShare,
		Records:       pa.Records,
	})

	if err != nil {
		return nil, err
	}

	return string(bs), nil
}

// MarshalJSON returns the JSON representation of a PoolAccount.
func (pa PoolAccount) MarshalJSON() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		return nil, err
	}

	return json.Marshal(poolAccountPretty{
		Address:       accAddr,
		PubKey:        "",
		AccountNumber: pa.AccountNumber,
		Id:            pa.Id,
		PoolParams:    pa.PoolParams,
		TotalWeight:   pa.TotalWeight,
		TotalShare:    pa.TotalShare,
		Records:       pa.Records,
	})
}

// UnmarshalJSON unmarshals raw JSON bytes into a PoolAccount.
func (pa *PoolAccount) UnmarshalJSON(bz []byte) error {
	var alias poolAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	pa.BaseAccount = authtypes.NewBaseAccount(alias.Address, nil, alias.AccountNumber, alias.Sequence)
	pa.Id = alias.Id
	pa.PoolParams = alias.PoolParams
	pa.TotalWeight = alias.TotalWeight
	pa.TotalShare = alias.TotalShare
	pa.Records = alias.Records

	return nil
}
