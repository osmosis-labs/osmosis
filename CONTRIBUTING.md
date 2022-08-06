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
        subdenom    string
        valid       bool
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
            // Create a denom
            res, err := suite.msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), tc.subdenom))
            if tc.valid {
                suite.Require().NoError(err)

                // Make sure that the admin is set correctly
                queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), & types.QueryDenomAuthorityMetadataRequest {
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
    testCases: = map[string] struct {
        expectedGenesis *types.GenesisState
    } {
        "default genesis": {
            expectedGenesis: types.DefaultGenesisState(),
        },
        "custom genesis": {
            expectedGenesis: customGenesis,
        },
    }

    for name, tc: = range testCases {
        suite.Run(name, func() {
            // Setup.
            app: = suite.App
            ctx: = suite.Ctx

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
func TestGetPoolAssetsByDenom(t * testing.T) {
    testCases: = map[string] struct {
        poolAssets                  []balancer.PoolAsset
        expectedPoolAssetsByDenom   map[string]balancer.PoolAsset
        err                         error
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

            require.Equal(t, tc.err, err)

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

* Run the [existing binary creation tool](https://github.com/osmosis-labs/osmosis/blob/main/.github/workflows/release.yml). Running `make -f contrib/images/osmobuilder/Makefile release` on the root of the repo will replicate the CI that creates the release folder containing the binaries.

* Make a PR to main, with a cosmovisor config, generated in tandem with the binaries from tool.
  * Should be its own PR, as it may get denied for Fork upgrades.

* Make a PR to main to update the import paths and go.mod for the new major release

* Should also make a commit into every open PR to main to do the same find/replace. (Unless this will cause conflicts)

* Do a PR if that commit has conflicts

* (Eventually) Make a PR that adds a version handler for the next upgrade
  * [Add v10 upgrade boilerplate #1649](https://github.com/osmosis-labs/osmosis/pull/1649/files)

* Update chain JSON schema's recommended versions in `chain.schema.json` located in the root directory.

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
