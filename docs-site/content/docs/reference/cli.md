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
