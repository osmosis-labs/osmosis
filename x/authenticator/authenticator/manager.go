package authenticator

type AuthenticatorManager struct {
	registeredAuthenticators []Authenticator
}

// NewAuthenticatorManager creates a new AuthenticatorManager.
func NewAuthenticatorManager() *AuthenticatorManager {
	return &AuthenticatorManager{
		registeredAuthenticators: []Authenticator{},
	}
}

// ResetAuthenticators resets all registered authenticators.
func (am *AuthenticatorManager) ResetAuthenticators() {
	am.registeredAuthenticators = []Authenticator{}
}

// InitializeAuthenticators initializes authenticators. If already initialized, it will not overwrite.
func (am *AuthenticatorManager) InitializeAuthenticators(initialAuthenticators []Authenticator) {
	if len(am.registeredAuthenticators) > 0 {
		return
	}
	am.registeredAuthenticators = initialAuthenticators
}

// RegisterAuthenticator adds a new authenticator to the list of registered authenticators.
func (am *AuthenticatorManager) RegisterAuthenticator(authenticator Authenticator) {
	am.registeredAuthenticators = append(am.registeredAuthenticators, authenticator)
}

// UnregisterAuthenticator removes an authenticator from the list of registered authenticators.
func (am *AuthenticatorManager) UnregisterAuthenticator(authenticator Authenticator) {
	for i, auth := range am.registeredAuthenticators {
		if auth.Type() == authenticator.Type() {
			am.registeredAuthenticators = append(am.registeredAuthenticators[:i], am.registeredAuthenticators[i+1:]...)
			break
		}
	}
}

// GetRegisteredAuthenticators returns the list of registered authenticators.
func (am *AuthenticatorManager) GetRegisteredAuthenticators() []Authenticator {
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
func (am *AuthenticatorManager) GetAuthenticatorByType(authenticatorType string) Authenticator {
	for _, authenticator := range am.GetRegisteredAuthenticators() {
		if authenticator.Type() == authenticatorType {
			return authenticator
		}
	}
	return nil
}
