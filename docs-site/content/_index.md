---
title: goninja
layout: hextra-home
---

<div class="hx-mt-6 hx-mb-6">
{{< hextra/hero-headline >}}
  Generate typed REST APIs&nbsp;<br class="sm:hx-block hx-hidden" />from annotated Go structs
{{< /hextra/hero-headline >}}
</div>

<div class="hx-mb-12">
{{< hextra/hero-subtitle >}}
  No reflection, no runtime magic — just plain Go files you can read, debug, and commit.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx-mb-6">
{{< hextra/hero-button text="Get Started" link="docs/getting-started" >}}
</div>

<div class="hx-mt-6"></div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Code-first"
    subtitle="The struct is the single source of truth. No separate schema files, no config YAML."
  >}}
  {{< hextra/feature-card
    title="Generated, not reflected"
    subtitle="goninja generate writes real .go files you commit. Errors show up at compile time, not in production."
  >}}
  {{< hextra/feature-card
    title="Plain net/http"
    subtitle="No framework lock-in for routing — generated code is a thin layer over the standard library."
  >}}
  {{< hextra/feature-card
    title="Safe by default"
    subtitle="Output schemas are always separate from your database model, so a sensitive field can't leak into a response just because it exists on the struct."
  >}}
  {{< hextra/feature-card
    title="Built to be extended"
    subtitle="Override any generated method, hook into create/update/delete, plug in your own auth — all without touching generated files."
  >}}
  {{< hextra/feature-card
    title="OpenAPI included"
    subtitle="Every generated resource emits an OpenAPI 3.0 fragment from the same annotations, served through Swagger UI or ReDoc."
  >}}
{{< /hextra/feature-grid >}}
