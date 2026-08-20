# Changelog

All notable changes to goninja are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

While goninja is pre-1.0, a minor version bump may include breaking changes.

Release notes with more context live at
[goninja.dev/docs/changelog](https://goninja.dev/docs/changelog/).

## [Unreleased]

### Changed

- **The generator now rejects models it cannot turn into working code**, instead
  of emitting a file that fails to compile. It reports every problem across
  every model in one run, names the file and field, and writes nothing when
  validation fails. Rejected: a missing `goninja`-tagged `ID` field, an `ID`
  typed anything but `int64` or `string`, a pointer relation field, `byid` on a
  non-relation field, and `filter` on a relation field.
- **An unknown `?order=` field is now a 400** rather than being silently
  ignored. Previously a typo returned unordered results with a 200, which is
  indistinguishable from a successful sort. The whitelist that makes ordering
  injection-safe is unchanged.

### Added

- `codegen.Validate`, and `Model.SourceFile` so a rejection can point at the
  file the struct was declared in.
- **`goninja.Unauthorized`**, a new framework error type mapped to 401 by
  `DefaultErrorMapper`. Every configured `Authenticator` declining a request
  now returns a proper JSON body (`{"code":"UNAUTHORIZED","error":"unauthorized"}`)
  through the same `Respond` path as every other error, instead of the plain
  `http.Error` text response it returned before.
- **An optional `Code` field on every framework error type**
  (`NotFound`, `ValidationError`, `BadRequest`, `Unauthorized`). Left unset,
  each type keeps its existing default JSON `"code"` (`NOT_FOUND`,
  `VALIDATION_FAILED`, `BAD_REQUEST`, `UNAUTHORIZED`); setting it lets a
  specific failure carry a more precise machine-readable identifier than the
  HTTP status alone provides, e.g. the unknown-`?order=`-field 400 above now
  sets `Code: "INVALID_ORDER_FIELD"`. All four types implement the new
  `goninja.CodedError` interface (`error` plus `ErrorCode() string`), which
  is what `DefaultErrorMapper` calls to resolve the body's `"code"` field.

## [0.1.0] - 2026-08-20

First pre-alpha release.

### Added

- **Code generation.** `goninja generate` reads a directory of annotated
  structs and writes one `<model>_generated.go` per model — separate
  `List`/`Retrieve`/`Create`/`Update` output types, HTTP handlers, GORM
  queries and an OpenAPI fragment, formatted with `go/format`.
- **Struct tag vocabulary.** Field-level `goninja:"..."` accepting `list`,
  `retrieve`, `create`, `update` and `filter`, plus the relation-only
  modifier `byid`.
- **Watch mode.** `goninja generate -watch` regenerates on any `.go` change
  under the models directory, debounced 300ms.
- **Relations without N+1.** `list` never preloads; `retrieve` preloads every
  relation field it carries. Belongs-to and has-many are both supported;
  `byid` exposes a related ID instead of nesting the full object.
- **Filtering, ordering and pagination.** `filter`-tagged fields become
  exact-match filters, with `_min`/`_max` range filters on numeric fields.
  Limit/offset pagination behind a `ListEnvelope[T]`
  (`{items, total, limit, offset}`), and `?order=-field` resolved against a
  per-model column whitelist.
- **Validation.** `validate` struct tags are copied onto `Create`/`Update`
  types only and checked before touching the database, returning a 422 keyed
  by JSON field name. `RegisterValidation` adds custom tags.
- **Error mapping.** `NotFound`, `ValidationError` and `BadRequest` map to
  404/422/400 through a pluggable `ErrorMapper`; anything else becomes a
  generic 500 that never leaks the underlying error.
- **Transactions.** `create`, `update` and `delete` handlers run inside
  `InTransaction`, so a failing hook rolls the whole operation back.
- **Hooks and overrides.** `BeforeCreateHook`, `AfterCreateHook`,
  `BeforeUpdateHook` and `BeforeDeleteHook`, method overrides via
  `SetSelf`/`Self()`, and `ResourceConfig`/`Configurer` for custom mount
  paths and restricted route sets.
- **Custom routes.** `Action` plus `BaseResource.SetActions` mount non-CRUD
  endpoints alongside the generated ones, documented in the same OpenAPI
  fragment.
- **Authentication.** `Authenticator` objects tried in order, with per-route
  policy through `Config`/`AuthPolicy` and `API.MountWithConfig`. The
  security schemes an authenticator describes are emitted into the generated
  OpenAPI document, so enforcement and documentation cannot drift apart.
  `HTTPBearer`, `HTTPBasic`, `APIKeyHeader` and `CookieKey` ship built in.
- **OpenAPI and docs UI.** `NewAPI`/`Mount` merge every resource's fragment
  into one document; `MountDocs` serves it as JSON alongside a rendered UI.
  Swagger UI and ReDoc are both vendored and embedded — no external CDN —
  behind the swappable `docsui.DocsUI` interface.
- **Testing helpers.** `goninjatest.NewDB` and `goninjatest.NewServer` drive
  a real generated resource over HTTP against in-memory SQLite, with no
  Postgres required.

### Known limitations

- GORM and `net/http` are assumed; there is no adapter layer for other ORMs
  or routers.
- A model's primary key must be a field literally named `ID`, typed `int64`
  or `string` (treated as a UUID).
- Relation fields must be a struct value or a slice of one — pointer
  relations are not supported.
- An unknown `?order=` field is ignored rather than rejected.
- No OpenAPI example values are generated.

[Unreleased]: https://github.com/caspel26/goninja/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/caspel26/goninja/releases/tag/v0.1.0
