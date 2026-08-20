---
title: CLI
weight: 3
---

The `goninja` CLI has a single command, `generate`, which parses a
directory of `goninja`-annotated model structs and writes one generated
file per model.

## Flags

| Flag | Type | Default | Notes |
|---|---|---|---|
| `-models` | string | `./models` | Directory containing goninja-annotated model structs. |
| `-out` | string | `./internal/api` | Output directory for generated code. |
| `-package` | string | `api` | Package name for generated code. |
| `-models-import` | string | *(empty)* | Import path of the models package, written verbatim into the generated import block. Required — the command errors if empty. |
| `-models-pkg` | string | `models` | Package name of the models package as used in Go source, used to qualify types (e.g. `models.Book`) in generated code. |
| `-watch` | bool | `false` | Watch `-models` and regenerate on change. |

`-models-import` and `-models-pkg` are separate flags because a Go
import path's last path segment does not always match the package's
declared name — goninja does not infer one from the other, so both must
be supplied explicitly when they differ.

## Example invocation

```shell
go run ./cmd/goninja generate \
  -models ./models \
  -out ./internal/api \
  -package api \
  -models-import github.com/you/yourapp/models \
  -models-pkg models
```

This reads every model struct under `./models`, and writes one
`<model>_generated.go` file per model into `./internal/api`, declared as
`package api`. Generated code imports the models package as
`github.com/you/yourapp/models`.

Place the output under `internal/` (as in the example above) so it is
enforced by the Go compiler as non-importable from outside your module —
generated code is meant to be used only through the app that owns it, not
imported by other modules.

Commit the generated output. See [Generated Code](../generated-code) for why.

## Validation

Before writing anything, the generator checks that every model it found can
actually be turned into working code. A model that cannot is rejected with a
message naming the file and field, and **no files are written** — so a bad
model never leaves a half-generated package behind.

```console
$ goninja generate -models-import myapp/models
goninja: codegen: models/book.go: Book: no goninja-tagged field named ID; every
  model needs one, typed int64 or string, and it must carry a goninja tag to be
  exposed (e.g. `goninja:"list,retrieve"`)
models/book.go: Book.Author: relation field is *Author; pointers are not
  supported, use a struct value for a belongs-to or a plain slice for a has-many
models/tag.go: Tag: ID is int, which the generator cannot use as a primary key;
  use int64 for a serial key or string for a UUID
$ echo $?
1
```

Every problem across every model is reported in a single run, rather than one
per attempt.

| Rejected | Why |
| --- | --- |
| No `goninja`-tagged field named `ID` | `Retrieve`, `Update` and `Delete` are all typed on the primary key |
| `ID` typed anything but `int64` or `string` | The path value is a string; `int64` is parsed with `strconv.ParseInt`, `string` is treated as a UUID |
| A pointer relation field (`*Author`, `[]*Book`) | Only a struct value or a plain slice of one is supported |
| `byid` on a field that is not a relation | The modifier only means something for a relation |
| `filter` on a relation field | A relation is not a column; tag its foreign key field instead |

This is the same guarantee the [tag reference](../tags/) describes from the
model's side: a mistake in a struct tag surfaces at generation time, with a
message about your model — not later as a compile error inside a
`DO NOT EDIT` file.

## go:generate

A `go:generate` directive lets `go generate ./...` regenerate without
remembering the full flag list:

```go {filename="models/generate.go"}
package models

//go:generate go run ../cmd/goninja generate -models . -out ../internal/api -package api -models-import github.com/you/yourapp/models
```

## Watch mode

Pass `-watch` to keep the process running and regenerate automatically
whenever a `.go` file under `-models` changes, instead of exiting after
the first generation.

- Only `.go` files trigger regeneration.
- Only the fsnotify `Write`, `Create`, and `Rename` operations trigger it —
  other operations (e.g. a bare chmod) do not.
- Changes are debounced exactly 300ms, implemented with `time.AfterFunc`,
  so a single editor save — often a write-to-temp-file-then-rename — triggers
  exactly one regeneration, not several.
- Ctrl+C (or SIGTERM) stops it cleanly.

```shell
go run ./cmd/goninja generate \
  -models ./models \
  -out ./internal/api \
  -package api \
  -models-import github.com/you/yourapp/models \
  -watch
```

{{< callout type="info" >}}
Watch mode is meant for local development. Run a plain, non-watch
`generate` in CI or as part of a build step, and commit the result — don't
rely on watch mode to keep committed generated code in sync.
{{< /callout >}}
