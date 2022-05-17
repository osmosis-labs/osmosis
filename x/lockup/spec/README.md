# `lockup`

## Abstract

Lockup module provides an interface for users to lock tokens into the
module to get incentives.

Users can lock tokens with specific duration and to unlock users should
start unlock and wait for the unlock period that's set initially. After
unlock period finish, users can claim tokens from the module.

This module provides interfaces for other modules to iterate the locks
efficiently and grpc query to check the status of locked coins.

## Contents

1. **[Concept](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Keeper](05_keeper.md)**\
6. **[Hooks](06_hooks.md)**\
7. **[Queries](07_queries.md)**\
8. **[Params](08_params.md)**
9. **[Endblocker](09_endblocker.md)**
