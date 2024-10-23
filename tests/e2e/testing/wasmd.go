package e2eTesting

import (
	"context"

	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmVmTypes "github.com/CosmWasm/wasmvm/v2/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ wasmKeeper.Messenger = (*MockMessenger)(nil)

// MockContractViewer mocks x/wasmd module dependency.
// Mock returns a contract info if admin is set.
type MockContractViewer struct {
	contractAdminSet map[string]string // key: contractAddr, value: adminAddr
	returnSudoError  error
}

// NewMockContractViewer creates a new MockContractViewer instance.
func NewMockContractViewer() *MockContractViewer {
	return &MockContractViewer{
		contractAdminSet: make(map[string]string),
		returnSudoError:  nil,
	}
}

// AddContractAdmin adds a contract admin link.
func (v *MockContractViewer) AddContractAdmin(contractAddr, adminAddr string) {
	v.contractAdminSet[contractAddr] = adminAddr
}

// GetContractInfo returns a contract info if admin is set.
func (v MockContractViewer) GetContractInfo(ctx context.Context, contractAddress sdk.AccAddress) *wasmdTypes.ContractInfo {
	adminAddr, found := v.contractAdminSet[contractAddress.String()]
	if !found {
		return nil
	}

	return &wasmdTypes.ContractInfo{
		Admin: adminAddr,
	}
}

// HasContractInfo returns true if admin is set.
func (v MockContractViewer) HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool {
	_, found := v.contractAdminSet[contractAddress.String()]
	return found
}

// Sudo implements the wasmKeeper.ContractInfoViewer interface.
func (v MockContractViewer) Sudo(ctx context.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	return nil, v.returnSudoError
}

func (v *MockContractViewer) SetReturnSudoError(returnSudoError error) {
	v.returnSudoError = returnSudoError
}

// MockMessenger mocks x/wasmd module dependency.
type MockMessenger struct{}

// NewMockMessenger creates a new MockMessenger instance.
func NewMockMessenger() *MockMessenger {
	return &MockMessenger{}
}

// DispatchMsg implements the wasmKeeper.Messenger interface.
func (m MockMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmVmTypes.CosmosMsg) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	return nil, nil, nil, nil
}
