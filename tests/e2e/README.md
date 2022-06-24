# End-to-end Tests

## Structure

### `e2e` Package

The `e2e` package defines an integration testing suite used for full
end-to-end testing functionality. This package is decoupled from
depending on the Osmosis codebase. It initializes the chains for testing
via Docker files. As a result, the test suite may provide the desired
Osmosis version to Docker containers during the initialization. This
design allows for the opportunity of testing chain upgrades in the
future by providing an older Osmosis version to the container,
performing the chain upgrade, and running the latest test suite.

The file e2e\_suite\_test.go defines the testing suite and contains the
core bootstrapping logic that creates a testing environment via Docker
containers. A testing network is created dynamically with 2 test
validators.

The file e2e\_test.go contains the actual end-to-end integration tests
that utilize the testing suite.

Currently, there is a single test in `e2e_test.go` to query the balances
of a validator.

## `initialization` Package

The `initialization` package introduces the logic necessary for initializing a
chain by creating a genesis file and all required configuration files
such as the `app.toml`. This package directly depends on the Osmosis
codebase.

## `upgrade` Package

The `upgrade` package starts chain initialization. In addition, there is
a Dockerfile `init-e2e.Dockerfile`. When executed, its container
produces all files necessary for starting up a new chain. These
resulting files can be mounted on a volume and propagated to our
production osmosis container to start the `osmosisd` service.

The decoupling between chain initialization and start-up allows to
minimize the differences between our test suite and the production
environment.

## `containers` Package

Introduces an abstraction necessary for creating and managing
Docker containers. Currently, validator containers are created
with a name of the corresponding validator struct that is initialized
in the `chain` package. 

## Running Locally

### To build chain initialization image

Please refer to `tests/e2e/initialization/README.md`

### To build the debug Osmosis image

```sh
    make docker-build-e2e-debug

### Environment variables

Some tests take a long time to run. Sometimes, we would like to disable them
locally or in CI. The following are the environment variables to disable
certain components of e2e testing.

- `OSMOSIS_E2E_SKIP_UPGRADE` - when true, skips the upgrade tests.
If OSMOSIS_E2E_SKIP_IBC is true, this must also be set to true because upgrade
tests require IBC logic.

- `OSMOSIS_E2E_SKIP_IBC` - when true, skips the IBC tests tests.

- `OSMOSIS_E2E_SKIP_CLEANUP` - when true, avoids cleaning up the e2e Docker
containers.

#### VS Code Debug Configuration

This debug configuration helps to run e2e tests locally and skip the desired tests.

```json
{
    "name": "E2E IntegrationTestSuite",
    "type": "go",
    "request": "launch",
    "mode": "test",
    "program": "${workspaceFolder}/tests/e2e",
    "args": [
        "-test.timeout",
        "30m",
        "-test.run",
        "IntegrationTestSuite",
        "-test.v"
    ],
    "env": {
        "OSMOSIS_E2E_SKIP_IBC": "true",
        "OSMOSIS_E2E_SKIP_UPGRADE": "true",
        "OSMOSIS_E2E_SKIP_CLEANUP": "true",
    }
}
```
