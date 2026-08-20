---
title: CLI Reference
weight: 4
---

```console
$ goninja generate \
    -models <dir> \
    -out <dir> \
    -package <name> \
    -models-import <full import path of the models package>
```

| Flag | Default | Description |
|---|---|---|
| `-models` | `./models` | Directory containing `goninja`-annotated model structs |
| `-out` | `./internal/api` | Output directory for generated code |
| `-package` | `api` | Package name for generated code |
| `-models-import` | *(required)* | Import path of the models package |
| `-models-pkg` | `models` | Package name of the models package, as used in Go source |
| `-watch` | `false` | Watch `-models` for changes and regenerate automatically |

## Watch mode

Pass `-watch` to keep the process running and regenerate automatically
whenever a `.go` file under `-models` changes. Changes are debounced
300ms so a single editor save — often a write-to-temp-file-then-rename —
triggers exactly one regeneration, not several. Ctrl+C (or SIGTERM)
stops it cleanly.

```console
$ goninja generate -models ./models -out ./internal/api \
    -package api -models-import github.com/you/yourapp/models -watch
goninja: generated 3 model(s) into ./internal/api
goninja: watching ./models for changes (Ctrl+C to stop)
```
