# Changelog

All notable changes to goninja are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

While goninja is pre-1.0, a minor version bump may include breaking changes.

Each entry names the package or file it touches and explains why the change
was made, not just what changed. Release notes with more context and worked
examples live at [goninja.dev/docs/changelog](https://goninja.dev/docs/changelog/).

## [Unreleased]

### Added

- **`API.SetErrorMapper`, `Config.DefaultErrorMapper`, and `goninja.NewErrorMapper`/`NewErrorMapping`** — `api.go`, `config.go`, `mapper.go`

  `API.SetErrorMapper(mappings ...ErrorMapping)` registers one or more
  per-error-type handlers on the app object itself, applied by both
  `Mount` and `MountWithConfig` to every resource that hasn't called its
  own `SetErrorMapper`, without requiring a `Config` built by hand just to
  reach `MountWithConfig`. It takes `ErrorMapping`s rather than a whole
  `ErrorMapper` so mappings from different files compose safely into one
  list: a plain `ErrorMapper` has no way to say "I didn't recognize this
  error, try the next one" (`DefaultErrorMapper` answers every error), so
  chaining whole `ErrorMapper`s would let an earlier one silently swallow
  everything after it — `ErrorMapping`'s own `Matches` avoids that.
  `Config.DefaultErrorMapper` still exists for setting a whole
  `ErrorMapper` explicitly on the `Config` passed to `MountWithConfig`
  (wins over `API.SetErrorMapper` when both are set — a resource's own
  `BaseResource.SetErrorMapper` still takes a whole `ErrorMapper` too, for
  full control at that scope). `NewErrorMapper`/`NewErrorMapping[T]` build
  an `ErrorMapping`/compose them into a plain `ErrorMapper`, matching via
  `errors.As` like `DefaultErrorMapper` itself, instead of a hand-written
  `MapError` switch. Resolution order per resource: its own
  `SetErrorMapper` wins if set, else `Config.DefaultErrorMapper`
  (explicit, or `API.SetErrorMapper`'s value), else the package
  `DefaultErrorMapper`.

## [0.3.1] - 2026-08-20

### Changed

- **`docsui.SpecSource` renamed to `docsui.SpecProvider`** — `docsui/docs.go`

  A SonarQube naming-convention fix (single-method interfaces should end
  in `-er`), matching the existing `openapi.OpenAPIProvider` precedent. A
  breaking rename for any caller referencing the type by name directly —
  pre-alpha, no compatibility shim.

- **`adapters/{gin,echo,chi}` test suites excluded from the duplication metric** — `sonar-project.properties`

  `adapter_test.go`/`docs_test.go` are intentionally near-identical across
  the three adapter modules (the same httptest suite written once and
  adapted per router, per each being a deliberately separate Go module
  with no shared test dependency) — this was tripping the
  `new_duplicated_lines_density` quality gate despite the actual adapter
  implementation code having 0% duplication. No code changes.

## [0.3.0] - 2026-08-20

### Added

- **Router adapters for gin, echo, and chi** — `router` (new), `adapters/gin`, `adapters/echo`, `adapters/chi` (new modules)

  `goninja.Resource.Register` now mounts on `goninja.Router`, a one-method
  interface `*http.ServeMux` already satisfies — plain `net/http` usage is
  unchanged. Each adapter translates a generated route's stdlib-style
  pattern (`"GET /books/{id}"`) into its router's own syntax and binds the
  matched path value back onto the request via `SetPathValue`, so a
  generated handler's `req.PathValue("id")` call needs no changes at all —
  only route registration is router-specific. Each adapter is its own Go
  module (own `go.mod`), so gin/echo/chi are never a dependency of a plain
  `net/http` project.

- **Benchmark suite** — `examples/prototype/benchmark_test.go`

  Three benchmarks (`go test -bench=.` via `make bench`) covering base
  list serialization, filter-clause building, and the automatic-`Preload`
  cost `Retrieve` pays on a relation field — the baseline for future
  optimization work.

- **Benchmark regression check in CI** — `scripts/bench-regression.sh`, `scripts/testdata/bench-baseline.txt`, `.github/workflows/bench.yml`

  `make bench-check` runs the benchmark suite (`-count=10`) and compares
  it against the committed baseline with `benchstat`, failing a PR if any
  benchmark's `sec/op`/`B/op`/`allocs/op` regresses by more than 25% and
  the difference is statistically significant (`benchstat`'s own `~`
  already filters out noise below that). The comparison table is also
  written to the GitHub Actions job summary, so a reviewer sees the actual
  numbers, not just pass/fail, and a self-contained HTML report
  (`reports/bench-report.html`, gitignored) is uploaded as a CI artifact
  for a more readable view than raw logs. `make bench-baseline` moves the
  baseline deliberately, after confirming a numbers change is expected.

- **`make bench-profile`** — `Makefile`

  Runs the benchmark suite with `-cpuprofile`/`-memprofile` and prints a
  top-10 `go tool pprof` summary for each, for finding what's actually
  worth optimizing rather than just whether something regressed.

### Changed

- **`Resource.Register` takes `goninja.Router` instead of `*http.ServeMux`** — `api.go`, `docsui/docs.go`, `internal/codegen/templates/model.go.tmpl`

  A breaking interface change for any hand-written `Resource` (not
  generated ones, which pick this up automatically on regeneration) —
  pre-alpha, no compatibility shim, same approach as the 0.2.0 auth
  redesign.

## [0.2.0] - 2026-08-20

### Added

- **`goninja.CodedError` and an optional `Code` field** — `errors.go`, `mapper.go`

  `NotFound`, `ValidationError`, `BadRequest`, and the new `Unauthorized`
  (below) implement `CodedError` (`error` plus `ErrorCode() string`). Left
  unset, each type keeps its existing default JSON `"code"` (`NOT_FOUND`,
  `VALIDATION_FAILED`, `BAD_REQUEST`, `UNAUTHORIZED`); setting `Code` lets a
  specific failure carry a more precise machine-readable identifier than the
  HTTP status alone provides — the unrecognized-`?order=`-field 400 below
  sets `Code: "INVALID_ORDER_FIELD"`.

- **`goninja.Unauthorized`** — `errors.go`, `resource.go`

  A new framework error type mapped to 401 by `DefaultErrorMapper`. Every
  configured `Authenticator` declining a request now returns a JSON body
  (`{"code":"UNAUTHORIZED","error":"unauthorized"}`) through the same
  `Respond` path as every other error, instead of the plain-text
  `http.Error` response it returned before.

- **`codegen.Validate`, `Model.SourceFile`** — `internal/codegen/validate.go`, `internal/codegen/ir.go`

  Lets a rejected model's error message point at the file the struct was
  declared in, alongside the change below.

### Changed

- **The generator rejects models it cannot turn into working code** — `internal/codegen/validate.go`, `internal/codegen/generate.go`

  Previously a bad model — no field named `ID`, an `ID` typed something
  other than `int64`/`string`, a pointer relation field, `byid` on a
  non-relation field, `filter` on a relation field — produced generated code
  that failed to compile, with nothing pointing back at the actual model
  that caused it. `Validate` now runs before any file is written, reports
  every problem across every model in one pass, and writes nothing at all
  when validation fails.

- **An unrecognized `?order=` field is now a 400** — `internal/codegen/templates/model.go.tmpl`

  Previously fell through silently and returned the default order with a
  200 — a response indistinguishable from a correctly sorted one. The
  column whitelist that makes ordering injection-safe is unchanged.

## [0.1.0] - 2026-08-20

First pre-alpha release.

### Added

- **Code generation** — `internal/codegen`, `internal/codegen/templates`

  `goninja generate` reads a directory of annotated structs and writes one
  `<model>_generated.go` per model — separate `List`/`Retrieve`/`Create`/
  `Update` output types, HTTP handlers, GORM queries and an OpenAPI
  fragment, formatted with `go/format`. The field-level tag vocabulary
  (`list`, `retrieve`, `create`, `update`, `filter`, plus the relation-only
  modifier `byid`) is a single struct tag, `ir.go`'s `Field.HasTag` the only
  place that interprets it.

- **Watch mode** — `cmd/goninja/watch.go`

  `goninja generate -watch` regenerates on any `.go` change under the
  models directory, debounced 300ms so one editor save produces one
  regeneration rather than several.

- **Relations without N+1** — `internal/codegen/templates/model.go.tmpl`

  `list` never preloads; `retrieve` is the detail view and preloads every
  relation field it carries — a guarantee of what the generator writes, not
  a default that can silently drift. Belongs-to and has-many relations are
  both supported; `byid` exposes a related ID instead of nesting the full
  object.

- **Filtering, ordering and pagination** — `pagination.go`, `internal/codegen/templates/model.go.tmpl`

  `filter`-tagged fields become exact-match filters on a generated
  `<Model>Filters` struct, with `_min`/`_max` range filters added for
  numeric fields. List responses are wrapped in `ListEnvelope[T]`
  (`{items, total, limit, offset}`), and `?order=-field` is resolved
  against a per-model column whitelist — the same whitelist that makes
  ordering injection-safe.

- **Validation** — `validate.go`

  `validate` struct tags are copied onto `Create`/`Update` types only —
  never onto read paths — and checked before the database is touched,
  returning a 422 keyed by JSON field name. `RegisterValidation` registers
  custom tags at startup.

- **Error mapping** — `errors.go`, `mapper.go`

  `NotFound`, `ValidationError` and `BadRequest` map to 404/422/400 through
  a pluggable `ErrorMapper`; anything else becomes a generic 500 that never
  leaks the underlying error message.

- **Transactions** — `resource.go`

  `create`, `update` and `delete` handlers run inside `InTransaction`, so a
  failing hook rolls the whole operation back — including `AfterCreate`
  rolling back the row just inserted.

- **Hooks and overrides** — `hooks.go`, `resource_config.go`, `resource.go`

  `BeforeCreateHook`, `AfterCreateHook`, `BeforeUpdateHook` and
  `BeforeDeleteHook`; method overrides via `SetSelf`/`Self()`; and
  `ResourceConfig`/`Configurer` for custom mount paths and restricted route
  sets.

- **Custom routes** — `actions.go`

  `Action` plus `BaseResource.SetActions` mount non-CRUD endpoints
  alongside the generated ones, documented in the same OpenAPI fragment as
  everything else on the resource.

- **Authentication** — `auth.go`, `authenticators.go`, `config.go`

  `Authenticator` objects tried in order, with per-route policy through
  `Config`/`AuthPolicy` and `API.MountWithConfig`. The security scheme an
  `Authenticator` describes is emitted into the generated OpenAPI document
  by the same resolution used to enforce it, so what's documented and
  what's enforced can't drift apart. `HTTPBearer`, `HTTPBasic`,
  `APIKeyHeader` and `CookieKey` ship built in.

- **OpenAPI and docs UI** — `api.go`, `openapi`, `docsui`

  `NewAPI`/`Mount` merge every resource's fragment into one document, and
  `MountDocs` serves it as JSON alongside a rendered UI. Swagger UI and
  ReDoc are both vendored and embedded — no external CDN — behind the
  swappable `docsui.DocsUI` interface.

- **Testing helpers** — `goninjatest`

  `goninjatest.NewDB` and `goninjatest.NewServer` drive a real generated
  resource over HTTP against in-memory SQLite, with no Postgres required.

### Known limitations

- GORM is assumed; there is no adapter layer for other ORMs. Routing was
  `net/http`-only with no adapter layer for other routers (fixed in
  [0.3.0]: gin, echo, and chi adapters).
- A model's primary key must be a field literally named `ID`, typed `int64`
  or `string` (treated as a UUID).
- Relation fields must be a struct value or a slice of one — pointer
  relations are not supported.
- An unknown `?order=` field is ignored rather than rejected (fixed in
  [0.2.0]).
- No OpenAPI example values are generated.

[Unreleased]: https://github.com/caspel26/goninja/compare/v0.3.1...HEAD
[0.3.1]: https://github.com/caspel26/goninja/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/caspel26/goninja/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/caspel26/goninja/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/caspel26/goninja/releases/tag/v0.1.0
