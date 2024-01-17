package authenticator

import (
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"
)

type AuthenticatorManager struct {
	registeredAuthenticators  []iface.Authenticator
	defaultAuthenticatorIndex int
}

// NewAuthenticatorManager creates a new AuthenticatorManager.
func NewAuthenticatorManager() *AuthenticatorManager {
	return &AuthenticatorManager{
		registeredAuthenticators:  []iface.Authenticator{},
		defaultAuthenticatorIndex: -1,
	}
}

// ResetAuthenticators resets all registered authenticators.
func (am *AuthenticatorManager) ResetAuthenticators() {
	am.registeredAuthenticators = []iface.Authenticator{}
}

// InitializeAuthenticators initializes authenticators. If already initialized, it will not overwrite.
func (am *AuthenticatorManager) InitializeAuthenticators(initialAuthenticators []iface.Authenticator) {
	if len(am.registeredAuthenticators) > 0 {
		return
	}
	am.registeredAuthenticators = initialAuthenticators
}

// RegisterAuthenticator adds a new authenticator to the list of registered authenticators.
func (am *AuthenticatorManager) RegisterAuthenticator(authenticator iface.Authenticator) {
	am.registeredAuthenticators = append(am.registeredAuthenticators, authenticator)
}

// UnregisterAuthenticator removes an authenticator from the list of registered authenticators.
func (am *AuthenticatorManager) UnregisterAuthenticator(authenticator iface.Authenticator) {
	for i, auth := range am.registeredAuthenticators {
		if auth == authenticator {
			am.registeredAuthenticators = append(am.registeredAuthenticators[:i], am.registeredAuthenticators[i+1:]...)
			break
		}
	}
}

// GetRegisteredAuthenticators returns the list of registered authenticators.
func (am *AuthenticatorManager) GetRegisteredAuthenticators() []iface.Authenticator {
	return am.registeredAuthenticators
}

// IsAuthenticatorTypeRegistered checks if the authenticator type is registered.
func (am *AuthenticatorManager) IsAuthenticatorTypeRegistered(authenticatorType string) bool {
	for _, authenticator := range am.GetRegisteredAuthenticators() {
		if authenticator.Type() == authenticatorType {
			return true
		}
	}
	return false
}

// GetAuthenticatorByType returns the base implementation of the authenticator type
func (am *AuthenticatorManager) GetAuthenticatorByType(authenticatorType string) iface.Authenticator {
	for _, authenticator := range am.GetRegisteredAuthenticators() {
		if authenticator.Type() == authenticatorType {
			return authenticator
		}
	}
	return nil
}

// SetDefaultAuthenticatorIndex sets the default authenticator index.
func (am *AuthenticatorManager) SetDefaultAuthenticatorIndex(index int) {
	am.defaultAuthenticatorIndex = index
	if am.defaultAuthenticatorIndex < 0 || am.defaultAuthenticatorIndex >= len(am.registeredAuthenticators) {
		panic("Invalid default authenticator index")
	}
}

// GetDefaultAuthenticator retrieves the default authenticator.
func (am *AuthenticatorManager) GetDefaultAuthenticator() iface.Authenticator {
	if am.defaultAuthenticatorIndex < 0 {
		// ToDo: Instead of panicking, maybe return a FalseAuthenticator that never authenticates?
		panic("Default authenticator not set")
	}
	return am.registeredAuthenticators[am.defaultAuthenticatorIndex]
}

type TransientStore struct {
	storeKey     storetypes.StoreKey
	transientCtx sdk.Context
}

func NewTransientStore(storeKey storetypes.StoreKey, ctx sdk.Context) *TransientStore {
	return &TransientStore{
		storeKey:     storeKey,
		transientCtx: ctx,
	}
}

func (as *TransientStore) ResetTransientContext(ctx sdk.Context) sdk.Context {
	as.transientCtx, _ = ctx.CacheContext()
	return as.transientCtx
}

func (as *TransientStore) GetKvStore() store.KVStore {
	return as.transientCtx.KVStore(as.storeKey)
}

func (as *TransientStore) GetTransientContext() sdk.Context {
	return as.transientCtx
}

func (as *TransientStore) GetTransientContextWithGasMeter(gasMeter sdk.GasMeter) sdk.Context {
	as.transientCtx = as.transientCtx.WithGasMeter(gasMeter)
	return as.transientCtx
}

func (as *TransientStore) WriteInto(ctx sdk.Context) {
	if as.transientCtx.IsZero() {
		panic("Transient context not set")
	}
	srcStore := as.transientCtx.KVStore(as.storeKey)
	destStore := ctx.KVStore(as.storeKey)
	syncStores(srcStore, destStore, true)
}

func (as *TransientStore) UpdateFrom(ctx sdk.Context) {
	if as.transientCtx.IsZero() {
		as.ResetTransientContext(ctx)
	}
	srcStore := ctx.KVStore(as.storeKey)
	destStore := as.transientCtx.KVStore(as.storeKey)
	syncStores(srcStore, destStore, true)
}

func syncStores(srcStore, destStore sdk.KVStore, clearDest bool) {
	// TODO: is there a cleaner way to do this? Ideally we'd just replace the entire store
	if clearDest {
		iterDest := destStore.Iterator(nil, nil)
		defer iterDest.Close()
		for ; iterDest.Valid(); iterDest.Next() {
			destStore.Delete(iterDest.Key())
		}
	}

	iterSrc := srcStore.Iterator(nil, nil)
	defer iterSrc.Close()
	for ; iterSrc.Valid(); iterSrc.Next() {
		destStore.Set(iterSrc.Key(), iterSrc.Value())
	}
}
