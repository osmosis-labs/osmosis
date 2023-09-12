package types

import "github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"

// NOTE: This should never return a pointer
// AUDIT: Check this function for security!
// AsAuthenticator converts an AccountAuthenticator to its corresponding Authenticator.
func (a *AccountAuthenticator) AsAuthenticator(
	am *authenticator.AuthenticatorManager,
) authenticator.Authenticator {
	for _, authenticatorCode := range am.GetRegisteredAuthenticators() {
		if authenticatorCode.Type() == a.Type {
			instance, err := authenticatorCode.Initialize(a.Data)
			if err != nil {
				return nil // ToDo: We should probably handle errors here
			}
			return instance
		}
	}
	return nil
}
