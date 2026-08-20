---
title: "Relations: Nested or by ID"
weight: 6
---

A relation field is nested as the related model's own `Retrieve` type by
default — the full object, `Preload`ed automatically:

```go
type Book struct {
    ID       string `gorm:"primaryKey;type:uuid" goninja:"list,retrieve"`
    AuthorID string `goninja:"list,retrieve,create,update,filter"`
    Author   Author `goninja:"retrieve"` // nested as {"author": {...full Author Retrieve...}}
}
```

Add `byid` to skip that — the field exposes only the related ID instead,
and its `Preload` never runs:

```go
Author Author `goninja:"retrieve,byid"` // {"author_id": "..."} — no nesting, no preload
```

Useful when a caller only ever needs the reference, not the full related
object, and the extra join/preload would be wasted work.
