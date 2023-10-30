# Sidecar Query Server

## Open Questions

- How to handle atomicity between ticks and pools? E.g. let's say a block is written between the time initial pools are read
and the time the ticks are read. Now, we have data that is partially up-to-date.