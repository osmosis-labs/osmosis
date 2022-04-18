# End-to-end Tests

The package e2e defines an integration testing suite used for full end-to-end
testing functionality.

The file e2e_suite_test.go defines the testing suite and contains the core
bootstrapping logic that creates a testing environment via Docker containers.
A testing network is created dynamically with 2 test validators.

The file e2e_test.go contains the actual end-to-end integration tests that
utilize the testing suite.

Currently, there is a single test in `e2e_test.go` to query the balances of a validator.
