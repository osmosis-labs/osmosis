//go:build !recoverylock_test

package recoverylock

// testBuild is false in every build without -tags recoverylock_test,
// including release builds. Such builds refuse test values.
const testBuild = false
