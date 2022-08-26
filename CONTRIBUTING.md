# Contributing

The following information provides a set of guidelines for contributing to the Osmosis chain dev repo. Use your best judgment, and, if you see room for improvement, please propose changes to this document.

The contributing guide for Osmosis explains the branching structure, how to use the SDK fork, how to make / test updates to SDK branches and how to create release notes.

Contributions come in the form of writing documentation, raising issues / PRs, and any other actions that help develop the Osmosis protocol documentation.

## First steps

The first step is to find an issue you want to fix. To identify issues we think are good for first-time contributors, we add the **good first issue** label.

We recommend setting up your IDE as per our [recommended IDE setup](https://docs.osmosis.zone/developing/osmosis-core/ide-guide.html) before proceeding.

If you have a feature request, please use the [feature-request repo](https://github.com/osmosis-labs/feature-requests). We also welcome you to [make an issue](https://github.com/osmosis-labs/osmosis/issues/new/choose) for anything of substance, or posting an issue if you want to work on it.

Once you find an existing issue that you want to work on or if you have a new issue to create, continue below.

## Proposing changes

To contribute a change proposal, use the following workflow:

1. [Fork the repository](https://github.com/osmosis-labs/osmosis).
2. [Add an upstream](https://docs.github.com/en/github/collaborating-with-pull-requests/working-with-forks/syncing-a-fork) so that you can update your fork.
3. Clone your fork to your computer.
4. Create a branch and name it appropriately.
5. Work on only one major change in one pull request.
6. Make sure all tests are passing locally.
7. Next, rince and repeat the following:

    1. Commit your changes. Write a simple, straightforward commit message. To learn more, see [How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/).
    2. Push your changes to your remote fork. To add your remote, you can copy/paste the following:
    ```sh

    #Remove origin

    git remote remove origin

    #set a new remote

    git remote add my_awesome_new_remote_repo [insert-link-found-in-source-subtab-of-your-repo]

    #Verify new remote

    git remote -v

    > my_awesome_new_remote_repo  [link-found-in-source-subtab-of-your-repo] (fetch)
    > my_awesome_new_remote_repo  [link-found-in-source-subtab-of-your-repo] (push)

    #Push changes to your remote repo

    git push <your_remote_name>

    #e.g. git push my_awesome_new_remote_repo
    ```
    3. Create a PR on the Osmosis repository. There should be a PR template to help you do so.
    4. Wait for your changes to be reviewed. If you are a maintainer, you can assign your PR to one or more reviewers. If you aren't a maintainer, one of the maintainers will assign a reviewer.
    5. After you receive feedback from a reviewer, make the requested changes, commit them to your branch, and push them to your remote fork again.
    6. Once approval is given, feel free to squash & merge!

## Writing tests

We use table-driven tests because they allow us to test similar logic on many different cases in a way that is easy to both implement and understand. [This article](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests) does a fantastic job explaining the motivation and structure of table-driven testing.

Making table-driven tests in an environment built on the Cosmos SDK has some quirks to it, but overall the structure should be quite similar to what is laid out in the article linked above.

We'll lay out three examples below (one that uses our format for messages, one that applies to keeper methods, and one that applies to our GAMM module), each of which will hopefully be simple enough to copy-paste into a test file and use as a starting point for your test-writing in the Osmosis Core repo.

### Generating unit tests using our Gotest template

To simplify (and speed up) the process of writing unit tests that fit our standard, we have put together a Gotest template that allows you to easily generate unit tests using built-in functionality for the Vscode Go plugin (complete with parameters listed, basic error checking logic etc.). The following two sections lay out how to generate a unit test automatically using this method.

#### 1. Setup
Note: this section assumes you already have the Go plugin for Vscode installed. Please refer to our [IDE setup docs](https://docs.osmosis.zone/developing/osmosis-core/ide-guide.html) if you haven't done any IDE setup yet.

Copy the `templates` folder into your `.vscode` folder from our main repo [here](https://github.com/osmosis-labs/osmosis/tree/main/.vscode). This folder has our custom templates for generating tests that fit our testing standards as accurately as possible.

Then, go to your `settings.json` file in your `.vscode` folder and add the following to it:

```go
    "go.generateTestsFlags": [
        "-template_dir",
        "[ABSOLUTE PATH TO TEMPLATES FOLDER]"
    ],
```
where `"[ABSOLUTE PATH TO TEMPLATES FOLDER]"` should look something like: `"User/ExampleUser/osmosis/.vscode/templates"`

#### 2. Generating a unit test
On the function you want to generate a unit test for, right click the function name and select `Go: Generate Unit Tests For Function`. This should take you to an automatically generated template test in the corresponding test file for the file your function is in. If there isn't a corresponding test file, it should automatically generate one, define its package and imports, and generate the test there.

### Rules of thumb for table-driven tests

1. Each test case should test one thing
2. Each test case should be independent from one another (i.e. ideally, reordering the tests shouldn't cause them to fail)
3. Functions should not be set as fields for a test case (this usually means that the table-driven approach is being sidestepped and that the logic in the function should probably be factored out to cover multiple/all test cases)
4. Avoid verbosity by creating local variables instead of constantly referring to struct field (e.g. doing `lockupKeeper := suite.App.LockupKeeper` instead of using `suite.App.LockupKeeper` every time).

### Example #1: [Message-Related Test]
This type of test is mainly for functions that would be triggered by incoming messages (we interact directly with the message server since all other metadata is stripped from a message by the time it hits the msg_server):

```go
func(suite *KeeperTestSuite) TestCreateDenom() {
    testCases := map[string] struct {
        subdenom            string
        expectError         bool
    } {

        "subdenom too long": {
            subdenom:   "assadsadsadasdasdsadsadsadsadsadsadsklkadaskkkdasdasedskhanhassyeunganassfnlksdflksafjlkasd",
            valid:      false,
        },
        "success case": {
            subdenom: "evmos",
			valid:    true,
        },
    }

    for name, tc := range testCases {
        suite.Run(name, func() {
            ctx := suite.Ctx
            msgServer := suite.msgServer
            queryClient := suite.queryClient

            // Create a denom
            res, err := msgServer.CreateDenom(sdk.WrapSDKContext(ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), tc.subdenom))
            
            if !tc.expectError {
                suite.Require().NoError(err)

                // Make sure that the admin is set correctly
                queryRes, err := queryClient.DenomAuthorityMetadata(ctx.Context(), & types.QueryDenomAuthorityMetadataRequest {
                    Denom: res.GetNewTokenDenom(),
                })

                suite.Require().NoError(err)
                suite.Require().Equal(suite.TestAccs[0].String(), queryRes.AuthorityMetadata.Admin)

            } else {
                suite.Require().Error(err)
            }
        })
    }
}
```
### Example #2: [Keeper-Related Test]
This type of test is mainly for functions that would be triggered by other modules calling public keeper methods (or just to unit-test keeper methods in general):

```go
// TestMintExportGenesis tests that genesis is exported correctly.
// It first initializes genesis to the expected value. Then, attempts
// to export it. Lastly, compares exported to the expected.
func(suite *KeeperTestSuite) TestMintExportGenesis() {
    testCases := map[string] struct {
        expectedGenesis *types.GenesisState
    } {
        "default genesis": {
            expectedGenesis: types.DefaultGenesisState(),
        },
        "custom genesis": {
            expectedGenesis: customGenesis,
        },
    }

    for name, tc := range testCases {
        suite.Run(name, func() {
            // Setup.
            app := suite.App
            ctx := suite.Ctx

            app.MintKeeper.InitGenesis(ctx, tc.expectedGenesis)

            // Test.
            actualGenesis := app.MintKeeper.ExportGenesis(ctx)

            // Assertions.
            suite.Require().Equal(tc.expectedGenesis, actualGenesis)
        })
    }
}
```

### Example #3: [Gamm-Related Test] 
Since the GAMM module is core to the Osmosis repo, it might be useful to have a good example of a well-structured GAMM-specific test. This example covers a simple getter function and validates the specific error messages around the function (as opposed to merely the presence of an error):

```go
func TestGetPoolAssetsByDenom(t *testing.T) {
    testCases := map[string] struct {
        poolAssets                          []balancer.PoolAsset
        expectedPoolAssetsByDenom           map[string]balancer.PoolAsset
        expectedErr                         error
    } {

        "one pool asset": {
            poolAssets: []balancer.PoolAsset {
                {
                    Token:  sdk.NewInt64Coin("uosmo", 1e12),
                    Weight: sdk.NewInt(100),
                },
            },
            expectedPoolAssetsByDenom: map[string]balancer.PoolAsset {
                "uosmo": {
                    Token:  sdk.NewInt64Coin("uosmo", 1e12),
                    Weight: sdk.NewInt(100),
                },
            },
        },

        "duplicate pool assets": {
            poolAssets: []balancer.PoolAsset {
                {
                    Token:  sdk.NewInt64Coin("uosmo", 1e12),
                    Weight: sdk.NewInt(100),
                }, {
                    Token:  sdk.NewInt64Coin("uosmo", 123),
                    Weight: sdk.NewInt(400),
                },
            },
            err: fmt.Errorf(balancer.ErrMsgFormatRepeatingPoolAssetsNotAllowed, "uosmo"),
        },
    }

    for name, tc := range testCases {
        t.Run(name, func(t *testing.T) {
            actualPoolAssetsByDenom, err := balancer.GetPoolAssetsByDenom(tc.poolAssets)

            require.Equal(t, tc.expectedErr, err)

            if tc.err != nil {
                return
            }

            require.Equal(t, tc.expectedPoolAssetsByDenom, actualPoolAssetsByDenom)
        })
    }
}
```

## Debug testing e2e locally

The e2e package defines an integration testing suite used for full end-to-end testing functionality. This package is decoupled from depending on the Osmosis codebase. It initializes the chains for testing via Docker files. 

As a result, the test suite may provide the desired Osmosis version to Docker containers during the initialization. This design allows for the opportunity of testing chain upgrades in the future by providing an older Osmosis version to the container, performing the chain upgrade, and running the latest test suite. 

The file `e2e_setup_test.go` defines the testing suite and contains the core bootstrapping logic that creates a testing environment via Docker containers. A testing network is created dynamically by providing the desirable number of validator configurations.

The file `e2e_test.go` contains the actual end-to-end integration tests that utilize the testing suite.


Additionally, there is an ability to disable certain components of the e2e suite. This can be done by setting the environment variables. See the [E2E test docs](https://github.com/osmosis-labs/osmosis/blob/main/tests/e2e/README.md)  or more details.

To get started:
- Run `make test-e2e`
- Inspect the logs of the docker containers and see if something it’s there
- `docker ps -a #` to list all docker containers
- Note the container id of the one you want to see the logs
- And then run `docker logs <CONTAINER_NAME>`  to debug via container logs

Please note that if the tests are stopped mid-way, the e2e framework might fail to start again due to duplicated containers. Make sure that
containers are removed before running the tests again: `docker containers rm -f $(docker containers ls -a -q)`.

Additionally, Docker networks do not get auto-removed. Therefore, you can manually remove them by running `docker network prune`.

## Working with the SDK

### Updating dependencies for builds

Vendor is a folder that go automatically makes if you run go mod vendor, which contains the source code for all of your dependencies. Its often helpful for local debugging. In order to update it...

Commit & push to the Cosmos-SDK fork in a new branch (see above steps for more details), and then you can grab the commit hash to do:

```sh
go get github.com/osmosis-labs/cosmos-sdk@{my commit hash}
```

You get something like:

```sh
go get: github.com/osmosis-labs/cosmos-sdk@v0.33.2 updating to
github.com/osmosis-labs/cosmos-sdk@v0.42.10-0.20210829064313-2c87644925da: parsing go.mod:
module declares its path as: github.com/cosmos/cosmos-sdk
but was required as: github.com/osmosis-labs/cosmos-sdk
```

Then you can copy paste the `v0.42.10-0.20210829064313-2c87644925da` part and replace the corresponding section of go.mod

Then do `go mod vendor`, and you're set.

### Changing things in vendor for local builds / local testing

In whichever folder you're running benchmarks for, you can test via:

`go test -benchmem -bench DistributionLogicLarge -cpuprofile cpu.out -test.timeout 30m -v`

Then once that is done, and you get the short benchmark results out, you can do:

`go tool pprof -http localhost:8080 cpu.out`

and take look at the graphviz output!

Note that if you are doing things that are low-level / small, the overhead of cpuprofile may mess with cache effects, etc. However for things like epoch code, or relatively large txs, this totally works!

### Branch structure of releases on v7, v6, v4

People still need those versions for querying old versions of the chain, and syncing a node from genesis, so we keep these updated!

For v6.x, and v4.x, most PRs to them should go to main and get a "backport" label. We typically use mergify for backporting. Backporting often takes place after a PR has been merged to main

### How to build proto files. (rm -rf vendor/ && make build-reproducible once docker is installed)

You can do rm -rf vendor and make build-reproducible to redownload all dependencies - this should pull the latest docker image of Osmosis. You should also make sure to do make proto-all to auto-generate your protobuf files. Makes ure you have docker installed.

If you get something like `W0503 22:16:30.068560 158 services.go:38] No HttpRule found for method: Msg.CreateBalancerPool` feel free to ignore that.

You can also feel free to do `make format` if you're getting errors related to `gofmt`. Setting this up to be [automatic](https://www.jetbrains.com/help/go/reformat-and-rearrange-code.html#reformat-on-save) for yourself is recommended.

## Major Release

There are several steps that go into a major release

* The GitHub release is created via this [GitHub workflow](https://github.com/osmosis-labs/osmosis/blob/main/.github/workflows/release.yml). The workflow is manually triggered from the [osmosis-ci repository](https://github.com/osmosis-labs/osmosis-ci). The workflow uses the `make build-reproducible` command to create the `osmosisd` binaries using the default [Makefile](https://github.com/osmosis-labs/osmosis/blob/main/Makefile#L99).

* Make a PR to main, with a cosmovisor config, generated in tandem with the binaries from tool.
  * Should be its own PR, as it may get denied for Fork upgrades.

* Make a PR to main to update the import paths and go.mod for the new major release

* Should also make a commit into every open PR to main to do the same find/replace. (Unless this will cause conflicts)

* Do a PR if that commit has conflicts

* (Eventually) Make a PR that adds a version handler for the next upgrade
  * [Add v10 upgrade boilerplate #1649](https://github.com/osmosis-labs/osmosis/pull/1649/files)

* Update chain JSON schema's recommended versions in `chain.schema.json` located in the root directory.

## Patch and Minor Releases

### Backport Flow

For patch and minor releases, we should already have
a release branch available in the repository. For example,
for any v11 release, we have a `v11.x` branch.

Therefore, for any change made to the `main` and, as long as it is
**state-compatible**, we must backport it to the last major release branch e.g.
`v11.x` when the next major release is v12.

This helps to minimize the diff of a major upgrade review process.

Additionally, it helps to safely and incrementally test
state-compatible changes by doing smaller patch releases. Contrary
to a major release, there is always an opportunity to safely downgrade
for any minor or patch release.

### State-compatibility

It is critical for the patch and minor releases to be state-machine compatible with
prior releases in the same major version. For example, v11.2.1 must be
compatible with v11.1.0 and v11.0.0.

This is to ensure **determinism** i.e. that given the same input, the nodes
will always produce the same output.

State-incompatibility is allowed for major upgrades because all nodes in the network
perform it at the same time. Therefore, after the upgrade, the nodes continue
functioning in a deterministic way.

#### Scope

The state-machine scope includes the following areas:

- All messages supported by the chain

- Transaction gas usage

- Whitelisted queries

- All `BeginBlock`/`EndBlock` logic

The following are **NOT* in the state-machine scope:

- Events

- Queries that are not whitelisted

- CLI interfaces

#### Validating State-Compatibility 

Tendermint ensures state compatibility by validating a number
of hashes that can be found here:
https://github.com/tendermint/tendermint/blob/9f76e8da150414ce73eed2c4f248947b657c7587/proto/tendermint/types/types.proto#L70-L77

Among the hashes that are commonly affected by our work and cause
problems are the `AppHash` and `LastResultsHash`. To avoid these problems, let's now examine how these hashes work.

##### App Hash

Cosmos-SDK takes an app hash of the state, and propagates it to Tendermint which,
in turn, compares it to the app hash of the rest of the network.
An app hash is a hash of hashes of every store's Merkle root.

For example, at height n, we compute:
`app hash = hash(hash(root of x/epochs), hash(root of  x/gamm)...)`

Then, Tendermint ensures that the app hash of the local node matches the app hash
of the network. Please note that the explanation and examples are simplified.

##### LastResultsHash

The `LastResultsHash` today contains:
https://github.com/tendermint/tendermint/blob/main/types/results.go#L47-L54

- Tx `GasWanted`

- Tx `GasUsed`

`GasUsed` being merkelized means that we cannot freely reorder methods that consume gas.
We should also be careful of modifying any validation logic since changing the
locations where we error or pass might affect transaction gas usage.

There are plans to remove this field from being Merkelized in a subsequent Tendermint release, at which point we will have more flexibility in reordering operations / erroring.

- Tx response `Data`

The `Data` field includes the proto message response. Therefore, we cannot
change these in patch releases.

- Tx response `Code`

This is an error code that is returned by the transaction flow. In the case of
success, it is `0`. On a general error, it is `1`. Additionally, each module
defines its custom error codes. For example, `x/mint` currently has the
following:
https://github.com/osmosis-labs/osmosis/blob/8ef2f1845d9c7dd3f422d3f1953e36e5cf112e73/x/mint/types/errors.go#L8-L10

As a result, it is important to avoid changing custom error codes or change
the semantics of what is valid logic in thransaction flows.

Note that all of the above stem from `DeliverTx` execution path, which handles:

- `AnteHandler`'s marked as deliver tx
- `msg.ValidateBasic`
- execution of a message from the message server

The `DeliverTx` return back to the Tendermint is defined [here][1].

#### Major Sources of State-incompatibility

##### Creating Additional State

By erroneously creating database entries that exist in Version A but not in
Version B, we can cause the app hash to differ across nodes running
these versions in the network. Therefore, this must be avoided.

##### Changing Proto Field Definitions

For example, if we change a field that gets persisted to the database,
the app hash will differ across nodes running these versions in the network.

Additionally, this affects `LastResultsHash` because it contains a `Data` field that is a marshaled proto message.


##### Returning Different Errors Given Same Input

Version A
```go
func (sk Keeper) validateAmount(ctx context.Context, amount sdk.Int) error {
    if amount.IsNegative() {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "amount must be positive or zero")
    }
    return nil
}
```

Version B
```go
func (sk Keeper) validateAmount(ctx context.Context, amount sdk.Int) error {
    if amount.IsNegative() || amount.IsZero() {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "amount must be positive")
    }
    return nil
}
```

Note that now an amount of 0 can be valid in "Version A". However, not in "Version B".
Therefore, if some nodes are running "Version A" and others are running "Version B",
the final app hash might not be deterministic.

Additionally, a different error message does not matter because it
is not included in any hash. However, an error code `sdkerrors.ErrInvalidRequest` does.
It translates to the `Code` field in the `LastResultsHash` and participates in
its validation.

##### Variability in Gas Usage

For transaction flows (or any other flow that consumes gas), it is important
that the gas usage is deterministic.

Currently, gas usage is being Merklized in the state. As a result, reordering functions
becomes risky.

Suppose my gas limit is 2000 and 1600 is used up before entering
`someInternalMethod`. Consider the following:

```go
func someInternalMethod(ctx sdk.Context) {
  object1 := readOnlyFunction1(ctx) # consumes 1000 gas
  object2 := readOnlyFunction2(ctx) # consumes 500 gas
  doStuff(ctx, object1, object2)
}
```
- It will run out of gas with `gasUsed = 2600` where 2600 getting merkelized
into the tx results.

```go
func someInternalMethod(ctx sdk.Context) {
  object2 := readOnlyFunction2(ctx) # consumes 500 gas
  object1 := readOnlyFunction1(ctx) # consumes 1000 gas
  doStuff(ctx, object1, object2)
}
```
- It will run out of gas with `gasUsed = 2100` where 2100 is getting merkelized
into the tx results.

Therefore, we introduced a state-incompatibility by merklezing diverging gas
usage.

#### Secondary Limitations To Keep In Mind

##### Network Requests to External Services

It is critical to avoid performing network requests to external services
since it is common for services to be unavailable or rate-limit.

Imagine a service that returns exchange rates when clients query its HTTP endpoint.
This service might experience downtime or be restricted in some geographical areas.

As a result, nodes may get diverging responses where some
get successful responses while others errors, leading to state breakage.

##### Randomness

Another critical property that should be avoided due to the likelihood
of leading the nodes to result in a different state.

##### Parallelism and Shared State

Threads and Goroutines might preempt differently in different hardware. Therefore,
they should be avoided for the sake of determinism. Additionally, it is hard
to predict when the multi-threaded state can be updated.

##### Hardware Errors

This is out of the developer's control but is mentioned for completeness.

### Pre-release auditing process

For every module with notable changes, we assign someone who was not a primary author of those changes to review the entire module.

Deliverables of review are:

* PR's with in-line code comments for things they had to figure out (or questions) 

* Tests / test comments needed to convince themselves of correctness 

* Spec updates

* Small refactors that helped in understanding / making code conform to consistency stds / improve code signal-to-noise ratio welcome

* (As with all PRs, should not be a monolithic PR that gets PR'd, even though that may be the natural way its first formed)

At the moment, we're looking for a tool that lets us statically figure out every message that had something in its code path that changed. Until a tool is found, we must do this manually.

We test in testnet & e2e testnet behaviors about every message that has changed

We communicate with various integrators if they'd like release-blocking QA testing for major releases
    * Chainapsis has communicated wanting a series of osmosis-frontend functionalities to be checked for correctness on a testnet as a release blocking item

[1]:https://github.com/cosmos/cosmos-sdk/blob/d11196aad04e57812dbc5ac6248d35375e6603af/baseapp/abci.go#L293-L303
