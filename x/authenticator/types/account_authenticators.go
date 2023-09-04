package types

// List of registered authenticators
// ToDo: This being global leads to issues with tests that use multiple apps (i.e.: ibctesting)
// Is it better to move this to state or to modify the tests to be able to handle it as-is?
var registeredAuthenticators []Authenticator
var defaultAuhenticatorIndex int = -1

func ResetAuthenticators() {
	registeredAuthenticators = []Authenticator{}
}

func InitializeAuthenticators(initialAuthenticators []Authenticator) {
	registeredAuthenticators = initialAuthenticators
}

// RegisterAuthenticator adds a new authenticator to the list of registered authenticators.
func RegisterAuthenticator(authenticator Authenticator) {
	registeredAuthenticators = append(registeredAuthenticators, authenticator)
}

// UnregisterAuthenticator removes an authenticator from the list of registered authenticators.
func UnregisterAuthenticator(authenticator Authenticator) {
	for i, auth := range registeredAuthenticators {
		if auth == authenticator { // assuming equality comparison works as intended for your authenticators
			// Remove the element at index i
			registeredAuthenticators = append(registeredAuthenticators[:i], registeredAuthenticators[i+1:]...)
			break
		}
	}
}

// GetRegisteredAuthenticators returns the list of registered authenticators.
func GetRegisteredAuthenticators() []Authenticator {
	return registeredAuthenticators
}

// IsAuthenticatorTypeRegistered returns true if the authenticator type is registered.
func IsAuthenticatorTypeRegistered(authenticatorType string) bool {
	for _, authenticator := range GetRegisteredAuthenticators() {
		if authenticator.Type() == authenticatorType {
			return true
		}
	}
	return false
}

func (a AccountAuthenticator) AsAuthenticator() Authenticator {
	for _, authenticator := range GetRegisteredAuthenticators() {
		if authenticator.Type() == a.Type {
			return authenticator
		}
	}
	return nil
}

func SetDefaultAuthenticatorIndex(index int) {
	defaultAuhenticatorIndex = index
	if defaultAuhenticatorIndex < 0 || defaultAuhenticatorIndex >= len(registeredAuthenticators) {
		panic("Invalid default authenticator index")
	}
}

func GetDefaultAuthenticator() Authenticator {
	if defaultAuhenticatorIndex < 0 {
		// ToDo: Instead of panicing maybe return a FalseAuthenticator that never authenticates?
		panic("Default authenticator not set")
	}
	return registeredAuthenticators[defaultAuhenticatorIndex]
}
