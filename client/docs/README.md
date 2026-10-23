# API documentation

Nodes with `api.swagger = true` serve Swagger UI at `/swagger/`. The page and
spec live in `static/swagger/` and are embedded into the binary through
`statik/statik.go`.

## Regenerating

From the repository root:

```sh
yarn --cwd client/docs install --frozen-lockfile
PATH="$PWD/client/docs/node_modules/.bin:$PATH" make docs
```

`make docs` downloads the pinned third-party protos, generates one Swagger file
per query service, merges them with `swagger-combine` into
`static/swagger/swagger.yaml`, and rebuilds `statik/statik.go`. The weekly
`Update Docs` workflow does the same and opens a PR.

## Adding a module

Add its generated file to `config.json`, with a tag rename so its endpoints get
their own group in the UI (every gRPC service is otherwise tagged `Query`):

```json
{
  "url": "./tmp-swagger-gen/osmosis/example/v1beta1/query.swagger.json",
  "operationIds": { "rename": { "Params": "ExampleParams" } },
  "tags": { "rename": { "Query": "osmosis.example.v1beta1", "Service": "osmosis.example.v1beta1" } }
}
```

`swagger-combine` fails on duplicate operation IDs; resolve them with
`operationIds.rename` as above.

## Swagger UI assets

`swagger-ui-bundle.js` and `swagger-ui.css` are vendored from
[`swagger-ui-dist`](https://www.npmjs.com/package/swagger-ui-dist) 5.28.1.
`yarn audit` does not cover them; update them by hand.
