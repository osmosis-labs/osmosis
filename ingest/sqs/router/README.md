# Router

## Trade-offs To Re-evaluate

- Router skips found route if token OUT is found in the intermediary
path by calling `validateAndFilterRoutes` function
- Router skips found route if token IN is found in the intermediary
path by calling `validateAndFilterRoutes` function
- In the above 2 cases, we could exit early instead of continuing to search for such routes
