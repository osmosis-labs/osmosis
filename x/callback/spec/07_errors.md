# Errors

The module exposes the following error codes which are used with the x/cwerrors module in case of callback failures.

```proto
enum ModuleErrors {
  // ERR_UNKNOWN is the default error code
  ERR_UNKNOWN = 0;
  // ERR_OUT_OF_GAS is the error code when the contract callback exceeds the gas limit allowed by the module
  ERR_OUT_OF_GAS = 1;
  // ERR_CONTRACT_EXECUTION_FAILED is the error code when the contract callback execution fails
  ERR_CONTRACT_EXECUTION_FAILED = 2;
}
```
