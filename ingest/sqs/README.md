# Sidecar Query Server Ingester

This is an implementation of the sidecar query server (SQS) ingester.

Please find the sidecar query server [here](https://github.com/osmosis-labs/sqs)

The use case for this is performing certain data and computationally intensive tasks outside of
the chain node or the clients. For example, routing falls under this category because it requires
all pool data for performing the complex routing algorithm.

SQS is meant to offload the query load from the chain node to a separate server. Primarily, we use it for swap routing.

## Integrator Guide

Follow [this link](https://hackmd.io/@3DOBr1TJQ3mQAFDEO0BXgg/S1bsqPAr6) to find a guide on how to 
integrate with the sidecar query server.
