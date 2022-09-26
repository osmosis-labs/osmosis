# Simulator

The simulator package aims to provide tooling that achieves:

* Long running state machine runs on randomized input
* The ability to assert relevant correctness properties hold for the state machine
* An API that is compatible with fuzz testing individual messages
* Assert (or identify breaks) in State machine compatability across versions

## Code Structure

This code is contending with legacy code, and is thus facing code abstraction boundaries that are being improved.

* executor - Logic for executing the simulator
* simtypes - Defines interfaces and public API's that other modules use for defining their code to interact with the simulator

Intended code structure:

* simtypes - API's that module writers need
* executor - API's needed to instantiate the simulator for its variety of usages
* executor stats - API's needed to debug executor runs
* executor types - Internal types for executor logic
