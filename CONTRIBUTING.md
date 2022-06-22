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

You can also feel free to do `make format` if you're getting errors related to `gofmt`.

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
