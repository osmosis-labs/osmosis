# API documentation tooling

Use Node.js 22.13+ (or Node.js 24+) and Yarn Classic:

```sh
yarn install --frozen-lockfile
yarn combine
yarn convert
yarn build
yarn test
yarn audit
```

`redoc-cli` is deprecated. Redocly CLI replaces its `bundle` command with
`build-docs`, removing the unpatched `braces` and `elliptic` dependency chains.

## Security pins

- Keep `@redocly/cli` at **2.33.2** for now. This patched, unbundled release lets
  Yarn resolve and audit its dependencies. Newer releases bundle dependencies;
  2.70.0 embeds DOMPurify 3.4.16, which is affected by the current alerts and
  cannot be upgraded with a Yarn resolution.
- The DOMPurify resolution requires 3.4.17+, fixing alerts
  [#722](https://github.com/osmosis-labs/osmosis/security/dependabot/722),
  [#723](https://github.com/osmosis-labs/osmosis/security/dependabot/723), and
  [#724](https://github.com/osmosis-labs/osmosis/security/dependabot/724).
- The Undici and Faker resolutions fix additional vulnerabilities in the CLI's
  pinned transitive dependencies. Remove them only after an upstream update
  and a clean audit.

The dependency audit does **not** cover prebuilt JavaScript inside packages or
remote scripts. Redocly normally emits a CDN script pointing to a Redoc browser
bundle with its own older sanitizer. `build-docs.mjs` replaces that script with
a local esbuild bundle of `redoc-browser.mjs` and the locked dependencies. The
result stays self-contained and uses DOMPurify 3.4.17 in the browser as well as
during server-side rendering. Tests verify the inline sanitizer version and
absence of remote scripts. A Chrome smoke test also verified hydration without
JavaScript errors or remote script requests.

The generated HTML and embedded Go assets have both been refreshed. After
future renderer updates, regenerate the Go asset from the repository root:

```sh
go run -mod=readonly github.com/rakyll/statik -src=client/docs/static -dest=client/docs -f -m
```

To test a build without changing committed HTML, use
`yarn build /tmp/osmosis-openapi.html`.
