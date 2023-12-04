# Ingest

This is a package that is responsible for ingesting end-of-block data into various
sinks. It is designed to be extensible. A user can add a new sink by implementing
an `Ingester` interface and then calling `RegisterIngester` in `app.go`.

Note that to avoid causing a chain halt, any error or panic occuring during ingestion
is logged and silently ignored.
