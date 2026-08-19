# goninja — Implementation Plan

Framework Go per generare API REST CRUD complete a partire da struct annotati: routing, validazione input/output, serializzazione, OpenAPI, filtri, paginazione. Ispirato nella filosofia a `django-ninja-aio-crud` (definisci il modello una volta, ottieni l'API), ma ripensato per essere idiomatico Go invece che una porta 1:1.

**Nome**: `goninja` come nome di lavoro. Esiste `avdept/goninja` (framework Rails-like, early development, apparentemente inattivo) — nessun conflitto sul path di import (`github.com/caspel26/goninja`), solo potenziale rumore nelle ricerche. Rivalutabile prima della pubblicazione.

---

## 1. Posizionamento

Verificato rispetto a quanto esiste oggi in Go:

| Progetto | Cosa fa | Differenza |
|---|---|---|
| **Huma / Fuego** | Genera OpenAPI dagli handler scritti a mano | Nessuna integrazione ORM, nessun CRUD automatico |
| **ent / GORM gen / sqlc** | Query type-safe generate da schema/modelli | Nessun layer REST/OpenAPI sopra |
| **nicolasbonnici/gorest** | Libreria runtime con generics; `processor` riduce il boilerplate | Devi scrivere 4 DTO a mano + converter + 5 handler per modello; vincolato a Fiber; config YAML obbligatoria; versioning imposto nelle URL |
| **ckoliber/gocrud** | CRUD automatico sopra Huma | Legato a Huma, meno controllo sulla generazione |

**Il nostro angolo**: partire dagli struct Go annotati (code-first), generare *tutto* il resto — DTO/schemi, handler, query, OpenAPI — con `net/http` standard e zero cerimonia di setup. Nessuno oggi copre esattamente questa combinazione.

---

## 2. Decisione strutturale: code generation

In Python, `ModelSerializer` costruisce schemi Pydantic a runtime via `create_model()`. Go non ha introspection equivalente a costo zero. Due opzioni:

- **Reflection a runtime** — più vicino allo spirito Python, ma overhead per richiesta, errori scoperti a runtime, poco idiomatico Go
- **Code generation** — un CLI scansiona gli struct e genera file `.go` da committare

**Scelta: code generation.** È l'approccio di `ent`, `sqlc`, `oapi-codegen` — tutti i progetti Go maturi in questo spazio. Errori a compile-time, zero overhead runtime, codice ispezionabile e debuggabile. Il costo (rigenerare dopo ogni modifica ai modelli) è mitigato dal watch mode.

---

## 3. Scope: cosa c'è in v0.1 e cosa no

Il piano è cresciuto molto in fase di design. Per evitare che resti un piano, v0.1 è deliberatamente ridotta:

**In v0.1**
- CRUD singolo (list, retrieve, create, update, delete)
- Schemi separati per list/retrieve/create/update
- Validazione input (tag `validate`) e serializzazione output tipizzata
- Filtri e paginazione
- OpenAPI + Swagger UI montati automaticamente
- Estensione via embedding, error mapper, hook pre/post
- Auth pluggable (middleware standard)
- FK e preload

**Rimandato a dopo v0.1 — non-goal espliciti**
- Bulk operations (`POST /books/bulk`)
- Soft delete
- Versionamento API
- M2M complesse
- RBAC / permessi a livello di campo
- Admin panel (mai)

---

## 4. Architettura

```
┌──────────────────────────────────────────────────┐
│  models/book.go — struct annotati (TU)             │
└────────────────────┬─────────────────────────────┘
                     │ goninja generate
┌────────────────────▼─────────────────────────────┐
│  internal/api/ — GENERATO, non modificare          │
│   book_schema.go    → List/Retrieve/Create/Update  │
│   book_resource.go  → logica CRUD (GORM)           │
│   book_handlers.go  → livello HTTP + validazione   │
│   book_routes.go    → registrazione route          │
│   book_openapi.go   → frammento OpenAPI            │
└────────────────────┬─────────────────────────────┘
                     │ embedding / interfacce
┌────────────────────▼─────────────────────────────┐
│  resources.go — le tue personalizzazioni (TU)      │
│   override metodi, hook, config, middleware        │
└────────────────────┬─────────────────────────────┘
                     │ api.Register(mux, ...)
┌────────────────────▼─────────────────────────────┐
│  goninja.API — aggregatore centrale                │
│   inietta DB + ErrorMapper, monta route, accumula  │
│   frammenti OpenAPI, serve /docs                   │
└──────────────────────────────────────────────────┘
```

---

## 5. Design delle API pubbliche

### 5.1 Il modello

```go
// models/book.go
type Book struct {
    ID        int64     `gorm:"primaryKey" goninja:"list,retrieve"`
    Title     string    `gorm:"size:120;not null" goninja:"list,retrieve,create,update" validate:"required,max=120"`
    AuthorID  int64     `goninja:"list,retrieve,create,update,filter"`
    Author    Author    `gorm:"foreignKey:AuthorID" goninja:"retrieve"`
    Price     float64   `goninja:"list,retrieve,create,update,filter" validate:"min=0"`
    Published bool      `goninja:"list,retrieve,create,update,filter"`
    Description string  `goninja:"retrieve,create,update"`
    Reviews   []Review  `gorm:"foreignKey:BookID" goninja:"retrieve"`
    CreatedAt time.Time `goninja:"list,retrieve"`
}
```

**`list` e `retrieve` sono tag distinti**: la lista resta leggera (niente `Description`, niente `Reviews`), il dettaglio è completo. Non è cosmetico — determina se la query fa il `Preload` o no.

### 5.2 Setup — contesto condiviso, niente ripetizioni

```go
func main() {
    db, _ := gorm.Open(postgres.Open(dsn))
    mux := chi.NewRouter()

    api := goninja.New(goninja.Config{
        DB:          db,                 // una volta sola
        ErrorMapper: myErrorMapper{},    // una volta sola
        Title:       "Bookstore API",
        Version:     "1.0.0",

        Middleware: []func(http.Handler) http.Handler{
            LoggingMiddleware(), CORSMiddleware(),
        },

        // Auth di default su tutte le risorse
        DefaultAuth: goninja.AuthPolicy{
            Protected:  []string{"create", "update", "delete"},
            Middleware: []func(http.Handler) http.Handler{JWTMiddleware(secret)},
        },
    })

    api.Register(mux, &generated.AuthorResource{})
    api.Register(mux, &MyBookResource{})
    api.MountDocs(mux, "/docs")

    http.ListenAndServe(":8080", mux)
}
```

`Register` inietta DB ed ErrorMapper nella `BaseResource` incorporata — l'utente non li passa mai.

### 5.3 Auth: default additivo, deroghe esplicite

**Principio di sicurezza**: la config per-risorsa può solo *aggiungere* protezione. Per rendere pubblica una route protetta di default serve un campo dedicato, così la scelta è sempre visibile nel codice della risorsa.

```go
func (r *MyBookResource) Config() goninja.ResourceConfig {
    return goninja.ResourceConfig{
        Path: "/books",
        Auth: goninja.AuthOverride{
            AlsoProtect: []string{"retrieve"},  // aggiunge
            Public:      []string{},             // rimuove SOLO se esplicito qui
        },
    }
}
```

Un `AlsoProtect` che dimentica `create` non lo rende pubblico — resta protetto dal default globale. È l'opposto del comportamento di una sovrascrittura piena, ed è deliberato.

### 5.4 Schemi generati

```go
// internal/api/book_schema.go — GENERATO
type BookListSchema struct {
    ID        int64     `json:"id"`
    Title     string    `json:"title"`
    AuthorID  int64     `json:"author_id"`
    Price     float64   `json:"price"`
    Published bool      `json:"published"`
    CreatedAt time.Time `json:"created_at"`
}

type BookRetrieveSchema struct {
    // ...tutti i campi di List, più:
    Description string         `json:"description"`
    Author      AuthorSchema   `json:"author"`
    Reviews     []ReviewSchema `json:"reviews"`
}

type BookCreateSchema struct {
    Title       string  `json:"title" validate:"required,max=120"`
    AuthorID    int64   `json:"author_id"`
    Price       float64 `json:"price" validate:"min=0"`
    Published   bool    `json:"published"`
    Description string  `json:"description"`
}

type BookUpdateSchema struct { /* campi di Create, tutti puntatori/opzionali */ }
```

Gli schemi di output sono **sempre** struct separate dal model GORM: impossibile esporre per sbaglio campi sensibili solo perché stanno nel model.

### 5.5 Resource generata

```go
// internal/api/book_resource.go — GENERATO
type BookResource struct {
    goninja.BaseResource // fornisce r.DB(ctx) e r.ErrorMapper()
}

func (r *BookResource) List(ctx context.Context, f BookFilters) ([]BookListSchema, int64, error) {
    var books []models.Book
    var total int64
    q := r.DB(ctx).Model(&models.Book{})

    if f.Published != nil { q = q.Where("published = ?", *f.Published) }
    if f.PriceMin != nil  { q = q.Where("price >= ?", *f.PriceMin) }
    if f.AuthorID != nil  { q = q.Where("author_id = ?", *f.AuthorID) }

    q.Count(&total) // per l'envelope di paginazione
    if err := q.Limit(f.Limit).Offset(f.Offset).Find(&books).Error; err != nil {
        return nil, 0, err
    }
    return toBookListSchemas(books), total, nil
}

func (r *BookResource) Retrieve(ctx context.Context, id int64) (*BookRetrieveSchema, error) {
    var book models.Book
    err := r.DB(ctx).Preload("Author").Preload("Reviews").First(&book, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, goninja.NotFound{Resource: "book", ID: id}
    }
    if err != nil { return nil, err }
    return toBookRetrieveSchema(&book), nil
}

// Create, Update, Delete: stesso schema
```

**Nota su N+1**: `List` non fa mai `Preload`, per costruzione — è il motivo per cui `list` e `retrieve` hanno tag separati. Documentato come garanzia del framework, non come dettaglio implementativo.

**`r.DB(ctx)` invece di `r.DB()`**: se il context contiene una transazione (vedi 5.7), restituisce quella; altrimenti la connessione base.

### 5.6 Livello HTTP generato — dove vive la validazione

```go
// internal/api/book_handlers.go — GENERATO
func (r *BookResource) createHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        var input BookCreateSchema
        if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
            goninja.Respond(w, r.ErrorMapper(), goninja.BadRequest{Detail: "JSON non valido"})
            return
        }

        // Validazione input dai tag `validate`
        if err := goninja.Validate(input); err != nil {
            goninja.Respond(w, r.ErrorMapper(), err) // → 422, errori per campo
            return
        }

        out, err := goninja.InTransaction(req.Context(), r.DB(req.Context()), func(ctx context.Context) (*BookRetrieveSchema, error) {
            if h, ok := any(r).(goninja.BeforeCreateHook[BookCreateSchema]); ok {
                if err := h.BeforeCreate(ctx, &input); err != nil { return nil, err }
            }
            created, err := r.Create(ctx, input)
            if err != nil { return nil, err }
            if h, ok := any(r).(goninja.AfterCreateHook[BookRetrieveSchema]); ok {
                if err := h.AfterCreate(ctx, created); err != nil { return nil, err }
            }
            return created, nil
        })

        if err != nil {
            goninja.Respond(w, r.ErrorMapper(), err)
            return
        }
        goninja.RespondJSON(w, 201, out) // serializzato secondo BookRetrieveSchema
    }
}
```

### 5.7 Transazioni

Hook + business logic girano dentro un'unica transazione. `r.DB(ctx)` dentro un hook restituisce la transazione, non la connessione base — così un `BeforeCreate` che scrive viene annullato se `Create` fallisce.

### 5.8 Utente autenticato nel context

Contratto minimo tra middleware auth e framework:

```go
// Il middleware auth (tuo) inserisce l'utente:
ctx = goninja.WithUser(ctx, myUser)

// La resource lo legge:
func (r *MyBookResource) List(ctx context.Context, f generated.BookFilters) ([]generated.BookListSchema, int64, error) {
    user, ok := goninja.UserFromContext(ctx)
    if ok && !user.IsAdmin {
        f.Published = ptr(true)
    }
    return r.BookResource.List(ctx, f)
}
```

`goninja.User` è un'interfaccia minima (solo `ID() string`) — il framework non impone una struttura utente.

### 5.9 Response envelope

Deciso ora perché cambiarlo dopo è breaking:

```json
// GET /books?limit=10&offset=0
{
  "items": [...],
  "total": 120,
  "limit": 10,
  "offset": 0
}
```

`retrieve`/`create`/`update` restituiscono l'oggetto nudo, senza envelope.

### 5.10 Personalizzazione via embedding

```go
// resources.go — file TUO, mai toccato dal generator
type MyBookResource struct {
    generated.BookResource
    cache *redis.Client
}

func (r *MyBookResource) Retrieve(ctx context.Context, id int64) (*generated.BookRetrieveSchema, error) {
    if cached, err := r.cache.Get(ctx, key(id)).Result(); err == nil {
        var b generated.BookRetrieveSchema
        json.Unmarshal([]byte(cached), &b)
        return &b, nil
    }
    book, err := r.BookResource.Retrieve(ctx, id) // riusa il default
    if err != nil { return nil, err }
    r.cache.Set(ctx, key(id), mustJSON(book), 5*time.Minute)
    return book, nil
}

// Validazione custom via hook
func (r *MyBookResource) BeforeCreate(ctx context.Context, in *generated.BookCreateSchema) error {
    var count int64
    r.DB(ctx).Model(&models.Author{}).Where("id = ?", in.AuthorID).Count(&count)
    if count == 0 {
        return goninja.ValidationError{Fields: map[string]string{"author_id": "autore non esistente"}}
    }
    return nil
}
```

**Perché funziona in Go**: il router chiama i metodi sul tipo concreto passato a `Register`, non su un tipo base che chiama sé stesso. Go non ha dispatch dinamico via embedding — se `BaseResource.List()` chiamasse internamente `r.Retrieve()`, risolverebbe sempre al metodo base. Qui non serve, perché è sempre il chiamante esterno a scegliere.

### 5.11 Error mapper

```go
type myErrorMapper struct{}

func (myErrorMapper) MapError(err error) (int, any) {
    var nf goninja.NotFound
    if errors.As(err, &nf) {
        return 404, map[string]string{"error": nf.Resource + " non trovato", "code": "NOT_FOUND"}
    }
    var ve goninja.ValidationError
    if errors.As(err, &ve) {
        return 422, map[string]any{"errors": ve.Fields, "code": "VALIDATION_FAILED"}
    }
    return 500, map[string]string{"error": "errore interno", "code": "INTERNAL"}
}
```

---

## 6. Fasi di implementazione

### Fase 0 — Prototipo / decision gate (12-16 hrs)

Un solo modello, persistenza in memoria, nessun ORM, un solo tag. Serve a rispondere a tre domande prima di impegnarsi:

- Il ciclo "modifico struct → rigenero → uso" è accettabile?
- Il codice generato è leggibile e debuggabile?
- Quanto è complesso mantenere i template man mano che i casi crescono?

**Exit criteria**: decisione scritta se procedere. Se il workflow è troppo scomodo, ci si ferma qui — 12-16 ore, non mesi.

---

### Fase 1 — Motore di code generation (20-28 hrs)

1. Parser con `go/parser` + `go/ast`: estrae struct e tag
2. IR intermedia disaccoppiata da parsing e output (riusata poi per OpenAPI)
3. Template con `text/template`
4. CLI `goninja generate`

**Exit criteria**: rigenera il prototipo Fase 0 tramite il motore vero, funziona su un secondo modello diverso.

---

### Fase 2 — GORM e query (14-18 hrs)

1. Generazione dei metodi Resource sopra GORM
2. `BaseResource` con `DB(ctx)` transaction-aware
3. Preload automatico in `Retrieve` dai tag `retrieve` su campi relazione
4. Migrations: delegate a `db.AutoMigrate()` di GORM — non le generiamo (scope creep evitato)

**Exit criteria**: CRUD end-to-end su Postgres reale.

---

### Fase 3 — Schemi, validazione, serializzazione (16-20 hrs)

1. Generazione dei 4 schemi distinti dai tag
2. Funzioni di conversione model ↔ schema
3. Validazione input con `go-playground/validator`
4. `ErrorMapper` + tipi errore del framework (`NotFound`, `ValidationError`, `BadRequest`)

**Exit criteria**: input non valido → 422 con dettaglio per campo; output mai contiene campi non dichiarati nello schema.

---

### Fase 4 — Filtri, paginazione, envelope (12-16 hrs)

1. `BookFilters` generato dai tag `filter`
2. Paginazione limit/offset con envelope
3. Ordinamento (`?order=-created_at`)

**Exit criteria**: `GET /books?published=true&price_min=10&order=-created_at&limit=20` funziona.

---

### Fase 5 — OpenAPI + Swagger montato (14-18 hrs)

1. Frammento OpenAPI generato per risorsa dalla stessa IR
2. `goninja.API` accumula i frammenti a ogni `Register`
3. `MountDocs` serve spec + Swagger UI via `go:embed` (nessuna CDN esterna)

**Exit criteria**: due risorse registrate, `/docs` le mostra entrambe. Spec verificato con `shimwire`.

---

### Fase 6 — Estensione: embedding, hook, config, auth (16-20 hrs)

1. Hook `BeforeCreate`/`AfterCreate`/`BeforeUpdate`/`BeforeDelete` come interfacce opzionali
2. Transazione che avvolge hook + operazione
3. `ResourceConfig` con path, route abilitate, `AuthOverride` additivo
4. `WithUser`/`UserFromContext`
5. `Config.DefaultAuth` e `Config.Middleware` globali

**Exit criteria**: risorsa con metodo sovrascritto, hook di validazione, path custom, auth selettiva — tutto sopravvive a `goninja generate`.

---

### Fase 7 — Testing e testabilità (10-14 hrs)

Non solo testare il framework: rendere testabili le risorse dell'utente, che è parte del prodotto.

1. Suite di test del framework (parser, generazione, validazione, transazioni)
2. `goninja.NewTestServer(resource)` per test end-to-end senza boilerplate `httptest`
3. Helper per DB di test (SQLite in memoria)

**Exit criteria**: un utente può testare la propria resource custom in meno di 10 righe.

---

### Fase 8 — Watch mode (4-6 hrs)

`goninja generate --watch` — rigenerazione automatica al salvataggio, mitiga l'attrito del codegen.

---

### Fase 9 — Documentazione e distribuzione (12-16 hrs)

- README con lo stesso standard di `shimwire`
- Progetto di esempio completo (sullo stile di `ninja-aio-blog-example`)
- Confronto onesto con gorest/gocrud/Huma nel README
- `go install`, release taggata

---

## 7. Tempistica

| Fase | Ore |
|---|---|
| 0 — Prototipo/decision gate | 12-16 |
| 1 — Code generation | 20-28 |
| 2 — GORM e query | 14-18 |
| 3 — Schemi/validazione | 16-20 |
| 4 — Filtri/paginazione | 12-16 |
| 5 — OpenAPI/Swagger | 14-18 |
| 6 — Estensione/hook/auth | 16-20 |
| 7 — Testing | 10-14 |
| 8 — Watch mode | 4-6 |
| 9 — Docs/distribuzione | 12-16 |

**Totale v0.1: 130-172 ore.** A 5-8 ore/settimana sono circa **5-7 mesi**. Le Fasi 0-5 (framework funzionante senza estensioni avanzate) sono ~90-116 ore, cioè 3-4 mesi.

Vale la pena essere espliciti: è un progetto lungo. La Fase 0 esiste per decidere presto se vale la pena, e la lista dei non-goal (sezione 3) per evitare che diventi il progetto infinito che abbiamo già scartato una volta.

---

## 8. Rischi

| Rischio | Mitigazione |
|---|---|
| Il codegen risulta troppo scomodo | Fase 0, decision gate a basso costo |
| Scope creep | Non-goal espliciti in sezione 3; migrations delegate a GORM |
| Progetto abbandonato a metà | Fasi 0-5 producono qualcosa di già utilizzabile; il resto è incrementale |
| Manutenzione oltre `shimwire` | Nessuna mitigazione tecnica — valutazione di priorità personale |
| Ecosistema si muove (nuovo concorrente) | È già successo durante la progettazione; il posizionamento code-first resta distinto |
