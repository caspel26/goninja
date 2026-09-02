---
title: Performance & Benchmarks
weight: 7
---

goninja keeps a small benchmark suite to catch regressions in the request
path. These are reproducible baselines for this repository, not claims that
one framework is universally faster than another.

## What is measured

Every benchmark seeds an in-memory SQLite database with 1,000 rows, starts a
real `goninjatest` HTTP server, then measures complete HTTP `GET` requests.
The measured work includes routing, generated handlers, GORM queries,
conversion to generated response types, JSON encoding, and response-body
draining. Seeding happens before the timer starts.

<div class="gn-benchmark-workload-grid">
  <article class="gn-benchmark-workload">
    <p class="gn-benchmark-workload-name">Task list</p>
    <code class="gn-benchmark-route">GET /tasks</code>
    <p>List query and collection serialization.</p>
  </article>
  <article class="gn-benchmark-workload">
    <p class="gn-benchmark-workload-name">Filtered book list</p>
    <code class="gn-benchmark-route">GET /books?published=true&amp;…</code>
    <p>Filter parsing, <code>WHERE</code> construction, count, list query, and serialization.</p>
  </article>
  <article class="gn-benchmark-workload">
    <p class="gn-benchmark-workload-name">Book detail + preload</p>
    <code class="gn-benchmark-route">GET /books/{id}</code>
    <p>Detail retrieval plus the automatic belongs-to preload.</p>
  </article>
</div>

{{< callout type="warning" >}}
This is not a production-deployment benchmark: there is no network hop,
reverse proxy, PostgreSQL server, application middleware, or
application-specific business logic. Do not compare these numbers directly
with a benchmark that measures a different stack or workload.
{{< /callout >}}

## Current tracked baseline

<section class="gn-benchmark-dashboard" aria-label="Tracked benchmark baseline">
  <div class="gn-benchmark-dashboard-header">
    <div>
      <p class="gn-benchmark-kicker">Tracked baseline</p>
      <p class="gn-benchmark-caption">Ten samples per benchmark · representative medians</p>
    </div>
    <a class="gn-benchmark-raw" href="https://github.com/caspel26/goninja/blob/main/scripts/testdata/bench-baseline.txt">View raw samples <span aria-hidden="true">↗</span></a>
  </div>
  <div class="gn-benchmark-grid">
    <article class="gn-benchmark-card">
      <p class="gn-benchmark-name">Task list</p>
      <code class="gn-benchmark-route">GET /tasks</code>
      <p class="gn-benchmark-time"><strong>75.5</strong><span>µs / op</span></p>
      <dl class="gn-benchmark-stats">
        <div><dt>Memory</dt><dd>19.6 KB</dd></div>
        <div><dt>Allocs</dt><dd>415</dd></div>
      </dl>
    </article>
    <article class="gn-benchmark-card">
      <p class="gn-benchmark-name">Filtered book list</p>
      <code class="gn-benchmark-route">GET /books?…</code>
      <p class="gn-benchmark-time"><strong>189.0</strong><span>µs / op</span></p>
      <dl class="gn-benchmark-stats">
        <div><dt>Memory</dt><dd>37.3 KB</dd></div>
        <div><dt>Allocs</dt><dd>840</dd></div>
      </dl>
    </article>
    <article class="gn-benchmark-card">
      <p class="gn-benchmark-name">Book detail + preload</p>
      <code class="gn-benchmark-route">GET /books/{id}</code>
      <p class="gn-benchmark-time"><strong>69.2</strong><span>µs / op</span></p>
      <dl class="gn-benchmark-stats">
        <div><dt>Memory</dt><dd>22.8 KB</dd></div>
        <div><dt>Allocs</dt><dd>276</dd></div>
      </dl>
    </article>
  </div>
</section>

The raw file, rather than this table, is the source of truth. Hardware, Go
version, database driver, operating-system load, and the benchmark workload
all affect the result.

## Run the suite

```console
$ make bench
```

That runs each benchmark once with Go's normal benchmark calibration. To
refresh the committed ten-sample baseline deliberately:

```console
$ make bench-baseline
```

Review and commit that file only when a change in the numbers is understood.
It is not an automatic maintenance task.

## Regression protection in CI

Pull requests that affect Go code, benchmark tooling, or their dependencies
run `make bench-check`. CI takes ten fresh samples, compares them with the
tracked baseline using `benchstat`, and fails only for a statistically
significant regression above 25% in time, bytes allocated, or allocations.
The comparison report is uploaded as a build artifact even when the check
fails.

The threshold is intentionally wide because shared CI runners are noisy. A
passing check means "no large measured regression under this workload," not
"performance is proven for every application." Profile a real workload
before making an optimization decision.
