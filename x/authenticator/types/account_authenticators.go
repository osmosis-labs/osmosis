package types

import (
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"
)

// NOTE: This should never return a pointer
// AUDIT: Check this function for security!
// AsAuthenticator converts an AccountAuthenticator to its corresponding Authenticator.
func (a *AccountAuthenticator) AsAuthenticator(
	am *authenticator.AuthenticatorManager,
) iface.Authenticator {
	for _, authenticatorCode := range am.GetRegisteredAuthenticators() {
		if authenticatorCode.Type() == a.Type {
			instance, err := authenticatorCode.Initialize(a.Data) // TODO: Pass the a.id
			if err != nil {
				return nil // TODO: We should probably handle errors here
			}
			return instance
		}
	}
	return nil
}
