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
performing the chain upgrade, and running the latest test suite. When
testing a normal upgrade, the e2e test suite submits an upgrade proposal at
an upgrade height, ensures the upgrade happens at the desired height, and
then checks that operations that worked before still work as intended. If
testing a fork, the test suite instead starts the chain a few blocks before
the set fork height and ensures the chain continues after the fork triggers
the upgrade. Note that a regular upgrade and a fork upgrade are mutually exclusive. 

The file e2e\_setup\_test.go defines the testing suite and contains the
core bootstrapping logic that creates a testing environment via Docker
containers. A testing network is created dynamically with 2 test
validators.

The file `e2e_test.go` contains the actual end-to-end integration tests
that utilize the testing suite.

Currently, there is a single IBC test in `e2e_test.go`.

Additionally, there is an ability to disable certain components
of the e2e suite. This can be done by setting the environment
variables. See "Environment variables" section below for more details.

## How It Works

Conceptually, we can split the e2e setup into 2 parts:

1. Chain Initialization

    The chain can either be initialized off of the current branch, or off the prior mainnet release and then upgraded to the current branch.

    If current, we run chain initialization off of the current Git branch
    by calling `chain.Init(...)` method in the `configurer/current.go`.

    If with the upgrade, the same `chain.Init(...)` function is run inside a Docker container
    of the previous Osmosis version, inside `configurer/upgrade.go`. This is
    needed to initialize chain configs and the genesis of the previous version that
    we are upgrading from.

    The decision of what configuration type to use is decided by the `Configurer`.
    This is an interface that has `CurrentBranchConfigurer` and `UpgradeConfigurer` implementations.
    There is also a `BaseConfigurer` which is shared by the concrete implementations. However,
    the user of the `configurer` package does not need to know about this detail.

    When the desired configurer is created, the caller may
    configure the chain in the desired way as follows:

    ```go
    conf, _ := configurer.New(..., < isIBCEnabled bool >, < isUpgradeEnabled bool >)

    conf.ConfigureChains()
    ```

    The caller (e2e setup logic), does not need to be concerned about what type of
    configurations is hapenning in the background. The appropriate logic is selected
    depending on what the values of the arguments to `configurer.New(...)` are.

    `CurrentBranchConfigurer` configures chains from current Git branch.
    `UpgradeConfigurer` takes older osmosis versions and configures chains from them. 

    The configurer constructor is using a factory design pattern
    to decide on what kind of configurer to return. Factory design
    pattern is used to decouple the client from the initialization
    details of the configurer. More on this can be found
    [here](https://www.tutorialspoint.com/design_pattern/factory_pattern.htm)

    The rules for deciding on the configurer type 
    are as follows:
    
    - If only `isIBCEnabled`, we want to have 2 chains initialized at the
    current branch version of Osmosis codebase

    - If only `isUpgradeEnabled`, that's invalid (we can decouple upgrade
     testing from IBC in a future PR)

    - If both `isIBCEnabled` and `isUpgradeEnabled`, we want 2 chain
    with IBC initialized at the previous Osmosis version

    - If none are true, we only need one chain at the current branch version
    of the Osmosis code

2. Setting up e2e components

    Currently, there exist the following components:

    - Base logic
        - This is the most basic type of setup where a single chain is created
        - It simply spins up the desired number of validators on a chain.
    - IBC testing
        - 2 chains are created connected by Hermes relayer
        - Upgrade Testing
        - 2 chains of the older Osmosis version are created, and
        connected by Hermes relayer
    - Upgrade testing
        - CLI commands are run to create an upgrade proposal and approve it
        - Old version containers are stopped and the upgrade binary is added
        - Current branch Osmosis version is spun up to continue with testing
    - State Sync Testing (WIP)
        - An additional full node is created after a chain has started.
        - This node is meant to state sync with the rest of the system.

    This is done in `configurer/setup.go` via function decorator design pattern
    where we chain the desired setup components during configurer creation.
    [Example](https://github.com/osmosis-labs/osmosis/blob/c5d5c9f0c6b5c7fdf9688057eb78ec793f6dd580/tests/e2e/configurer/configurer.go#L166)

## `initialization` Package

The `initialization` package introduces the logic necessary for initializing a
chain by creating a genesis file and all required configuration files
such as the `app.toml`. In addition, it starts chain initialization.
This package directly depends on the Osmosis codebase.

In addition, there is a Dockerfile `init-e2e.Dockerfile`.
When executed, its container produces all files necessary for starting up a new chain. These
resulting files can be mounted on a volume and propagated to our
production osmosis container to start the `osmosisd` service.

## `containers` Package

Introduces an abstraction necessary for creating and managing
Docker containers. Currently, validator containers are created
with a name of the corresponding validator struct that is initialized
in the `initialization/chain` package.

## Running From Current Branch

### To build chain initialization image

Please refer to `tests/e2e/initialization/README.md`

### To build the debug Osmosis image

```sh
    make docker-build-debug
```

### Environment variables

Some tests take a long time to run. Sometimes, we would like to disable them
locally or in CI. The following are the environment variables to disable
certain components of e2e testing.

- `OSMOSIS_E2E_SKIP_UPGRADE` - when true, skips the upgrade tests.
If OSMOSIS_E2E_SKIP_IBC is true, this must also be set to true because upgrade
tests require IBC logic.

- `OSMOSIS_E2E_SKIP_IBC` - when true, skips the IBC tests.

- `OSMOSIS_E2E_SKIP_STATE_SYNC` - when true, skips the state sync tests.

- `OSMOSIS_E2E_SKIP_CLEANUP` - when true, avoids cleaning up the e2e Docker
containers.

- `OSMOSIS_E2E_FORK_HEIGHT` - when the above "IS_FORK" env variable is set to true, this is the string
of the height in which the network should fork. This should match the ForkHeight set in constants.go

- `OSMOSIS_E2E_UPGRADE_VERSION` - string of what version will be upgraded to (for example, "v10")

- `OSMOSIS_E2E_DEBUG_LOG` - when true, prints debug logs from executing CLI commands
via Docker containers. Set to true in CI by default.

#### VS Code Debug Configuration

This debug configuration helps to run e2e tests locally and skip the desired tests.

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "E2E: (make test-e2e-short)",
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
            "buildFlags": "-tags e2e",
            "env": {
                "OSMOSIS_E2E": "True",
                "OSMOSIS_E2E_SKIP_IBC": "true",
                "OSMOSIS_E2E_SKIP_UPGRADE": "true",
                "OSMOSIS_E2E_SKIP_CLEANUP": "true",
                "OSMOSIS_E2E_SKIP_STATE_SYNC": "true",
                "OSMOSIS_E2E_UPGRADE_VERSION": "v13",
                "OSMOSIS_E2E_DEBUG_LOG": "true",
            },
            "preLaunchTask": "e2e-setup"
        }
    ]
}
```

### Common Problems

Please note that if the tests are stopped mid-way, the e2e framework might fail to start again due to duplicated containers. Make sure that
containers are removed before running the tests again: `docker containers rm -f $(docker containers ls -a -q)`.

Additionally, Docker networks do not get auto-removed. Therefore, you can manually remove them by running `docker network prune`.


## Good to know things about e2e

This section contains common "gotchas" that is sometimes very good to know when working with e2e.

1. Order of executions of tests in `e2e_test.go` is alphabetic.

    If one test fails and breaks the chain, all subsequent tests might also fail. It is important to know about an order, because sometimes you might see a failed 
    test in `e2e_test.go`, however, in reality, the core reason of failure lays in something else.

    Example: you see `TestAddToExistingLock` failing. It might be because of something failing in test, of course. However, you need to keep in mind that something potentially
    broke during the `e2e` setup and because of that, lexicographically first test (currently `TestAddToExistingLock`) fails.

    A way to deal with this problem: disable upgrade logic by setting `OSMOSIS_E2E_SKIP_UPGRADE` to `false` and see how test performs.

2. Node update failure before running `e2e_test.go`

    Nodes can completely crash when upgrading for many reasons. 
    To detect that, we check the node's `Status` as soon as the upgrade is complete.
    An active node should return a response and no other errors. If not, all test would not be ran.

    To fix that, check the docker container and see if you node is working. Or remove all related container then rerun e2e test

3. Transaction fees.

    Consensus min fee is set to "0.0025". This is how much is charged for 1 unit of gas. By default, we specify the gas limit
    of `40000`. Therefore, each transaction should take `fee = 0.0025 uosmo / gas * 400000 gas = 1000 fee token.
    The fee denom is set in `tests/e2e/initialization/config.go` by the value of `E2EFeeToken`.
    See "Hermes Relayer - Consensus Min Fee" section for the relevant relayer configuration.

## Hermes Relayer

The configuration is built using the script in `tests/e2e/scripts/hermes_bootstrap.sh`

Repository: <https://github.dev/informalsystems/hermes>

### Consensus Min Fee

We set the following parameters Hermes configs to enable the consensus min fee in Osmosis:

    - `gas_price` - Specifies the price per gas used of the fee to submit a transaction and
    the denomination of the fee. The specified gas price should always be greater or equal to the `min-gas-price`
    configured on the chain. This is to ensure that at least some minimal price is 
    paid for each unit of gas per transaction.
    In Osmosis, we set consensus min fee = .0025 uosmo / gas * 400000 gas = 1000
    See ConsensusMinFee in x/txfees/types/constants.go

    - `default_gas` - the gas amount to use when simulation fails.

## Debugging

This section contains information about debugging osmosis's `e2e` tests.

1. Executing commands on docker containers

    Currently, `e2e` emits a lot of useful logs into standard output. However, sometimes that might not be enough. In this case, it is a good practice to run some commands inside docker containers and inspect outputs.
    In order to do that, run:
        
    ```sh
        docker exec < container name/id > < command >
    ```

    This will execute the specified command and print the response to standard output. 
    Example: `docker exec osmo-test-a-node-prune-nothing-snapshot osmosisd status` will print a node status of `osmo-test-a-node-prune-nothing-snapshot` container. 

2. Viewing docker container logs

    Another useful thing to do when debugging some low level error is inspecing container's logs. This can be done by running:

    ```sh
        docker logs < container name/id >
    ```

    Example: `docker logs osmo-test-a-node-prune-nothing-snapshot` will print logs emitted by container `osmo-test-a-node-prune-nothing-snapshot` to your standard output.

3. Viewing docker container logs when run osmosis's `e2e` tests in CI

    When a failure occurs in CI, we have persists these log files of all container running as artifacts `logs.tgz.`