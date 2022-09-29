# Simulator

The simulator package aims to provide tooling that achieves:

* Long running state machine runs on randomized input
* The ability to assert relevant correctness properties hold for the state machine
* An API that is compatible with fuzz testing individual messages
* Assert (or identify breaks) in State machine compatability across versions

## State initialization

Currently the initialization of the simulator is a mess of spaghetti code / broken abstractions, inherited from the SDK.

We are iteratively cleaning up more and more of this.

The direction we should be moving towards is:
* The simulator executor can start a chain in one of two forms:
  * From a RequestInitChain
  * From a state snapshot (e.g. IAVL DB's from a live chain)
* The simulator instantiator can then:
  * Create a RequestInitChain from
    * A genesis file
    * A custom parameterization file + randomization
    * Randomizing genesis generation
    * A mix of the above
  * Create a copy of a state snapshot

And then we have tooling to ease the simulator instantiator's burden in creating the request init chain.
(There is some work in this direction, e.g. per module randomize genesis or use default genesis if unimplemented)

## Code Structure

This code is contending with legacy code, and is thus facing code abstraction boundaries that are being improved.

* executor - Logic for executing the simulator
* simtypes - Defines interfaces and public API's that other modules use for defining their code to interact with the simulator

Intended code structure:

* simtypes - API's that module writers need
* executor - API's needed to instantiate the simulator for its variety of usages
* executor/stats - API's needed to debug executor runs
* executor/types - Internal types for executor logic
