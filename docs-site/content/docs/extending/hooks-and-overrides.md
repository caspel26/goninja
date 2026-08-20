---
title: Hooks and Overrides
weight: 1
---

```go
// handlers/author.go — your file, never touched by the generator.
package handlers

import (
    "context"
    "log"

    "gorm.io/gorm"

    "github.com/caspel26/goninja"
    "myapp/internal/api"
)

type authorWithAudit struct {
    *api.AuthorResource
}

// BeforeCreateHook: runs inside the same transaction as Create, and an
// error here rolls the whole request back — nothing gets written.
func (r *authorWithAudit) BeforeCreate(ctx context.Context, in *api.AuthorCreate) error {
    if in.Name == "" {
        return goninja.ValidationError{Fields: map[string]string{"name": "required"}}
    }
    return nil
}

// AfterCreateHook: runs once the row exists, still inside that transaction.
func (r *authorWithAudit) AfterCreate(ctx context.Context, out *api.AuthorRetrieve) error {
    log.Printf("author created: %s", out.ID)
    return nil
}

func NewAuthorWithAudit(db *gorm.DB) *authorWithAudit {
    inner := api.NewAuthorResource(db)
    w := &authorWithAudit{AuthorResource: inner}
    inner.SetSelf(w) // wires the wrapper's hooks/overrides into the generated handlers
    return w
}
```

`BeforeCreateHook`/`AfterCreateHook`/`BeforeUpdateHook`/`BeforeDeleteHook`
are plain optional interfaces (`goninja` root package) — implement only
the ones you need. The same `SetSelf` wiring also makes an overridden
`Retrieve`/`List`/`Update`/`Delete` method take effect, e.g. to add
caching in front of the generated query without forking it.
