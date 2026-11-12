//go:build recoverylock_test

package recoverylock

// testBuild is true only when compiled with -tags recoverylock_test. Only such
// builds accept test values and the known test addresses.
const testBuild = true
