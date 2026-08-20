---
title: Testing Your Own Resources
weight: 7
---

`goninjatest` (a separate module-level package, like `openapi`/`docsui`/`id`
— only test code needs to import it) gets a resource from zero to a real
HTTP request in a few lines: `NewDB` opens and `AutoMigrate`s an in-memory
SQLite database, `NewServer` mounts any number of resources on a fresh
`httptest.Server`. Both clean up after themselves via `t.Cleanup`.

```go
func TestBookResource_Create(t *testing.T) {
    db := goninjatest.NewDB(t, &models.Book{}, &models.Author{})
    srv := goninjatest.NewServer(t, api.NewBookResource(db))

    resp, err := http.Post(srv.URL+"/books", "application/json",
        strings.NewReader(`{"title":"Dune","author_id":"...","price":9.99}`))
    if err != nil || resp.StatusCode != http.StatusCreated {
        t.Fatalf("POST /books: err=%v status=%v", err, resp.StatusCode)
    }
}
```

No Postgres, no hand-rolled `httptest.NewServer(mux)` boilerplate, no
`gorm.Open` call to remember the driver for — generated resource code has
no Postgres-specific behavior, so the same in-memory SQLite connection
exercises it end to end.
