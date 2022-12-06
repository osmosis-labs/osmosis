# Downtime-detector

For several use cases, we need a module that can detect when the chain is recovering from downtime. We want to be able to efficiently know "Has it been $RECOVERY_PERIOD minutes since the chain has been down for $DOWNTIME_PERIOD", and expose this as a query to contracts.

So for instance, you'd want to know if it has been at least 10 minutes, since the chain was down for > 30 minutes. Since you assume in such an event that it may take ~10 minutes for price oracles to be arb'd to correct.
Suggested Design

Theres a couple designs, such as:

* Iterating over block times from the last N blocks (with a heuristic filter based on average block time)
    * Implies bounds on recovery time
    * Linear iteration if heuristic is met
    * Requires encoding expected block time
* Restricting downtime period, and storing a state entry for last time a downtime of length $D occurred

Because this will be in important txs for contracts, we need to go with the approach that has minimal query compute, which is the latter. So we explain that in more depth.

We restrict the $DOWNTIME_PERIOD options that you can query, to be: 30seconds, 1 min, 2 min, 3 min, 4 min, 5 min, 10 min, 20 min, 30 min, 40 min, 50 min, 1 hr, 1.5hr, 2 hr, 2.5 hr, 3 hr, 4 hr, 5 hr, 6 hr, 9hr, 12hr, 18hr, 24hr, 36hr, 48hr.

In the downtime detector module, we store state entries for:

* Last blocks timestamp
* For each period, last time there was downtime

Then in every begin block:

* Store last blocks timestamp
* if time since last block timestamp >= 30 seconds, iterate through all $DOWNTIME_PERIODS less than the downtime, and in each add a state entry for the current block time

Then our query for has it been $RECOVERY_PERIOD since $DOWNTIME_PERIOD, simply reads the state entry for that $DOWNTIME_PERIOD, and then checks if time difference between now and that block is > RECOVERY_PERIOD.