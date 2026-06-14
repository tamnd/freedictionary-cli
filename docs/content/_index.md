---
title: "freedictionary"
description: "A command line for the Free Dictionary API."
heroTitle: "Word definitions, from the command line"
heroLead: "A command line for api.dictionaryapi.dev. One pure-Go binary, no API key, output that pipes into the rest of your tools, and a resource-URI driver other programs can address."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

`freedictionary` looks up word definitions from the public Free Dictionary API.
No API key required. Supports English and 11 other languages.

```bash
freedictionary define hello                  # all meanings for "hello"
freedictionary define hello -o json          # as JSON, ready for jq
freedictionary define bonjour --lang fr      # French
freedictionary define hello -o jsonl | jq .  # pipe into jq
```

There is nothing to sign up for and nothing to run alongside it. Output adapts
to where it goes: an aligned table on your terminal, JSONL the moment you pipe
it somewhere.

## Two ways to use it

- **As a command** for looking up words by hand or in a script. Start with
  the [quick start](/getting-started/quick-start/).
- **As a resource-URI driver** so a host like
  [ant](https://github.com/tamnd/ant) can address word definitions as
  `freedictionary://` URIs. See
  [resource URIs](/guides/resource-uris/).

Both are the same code: one operation, declared once, is a CLI command, an HTTP
route, an MCP tool, and a URI dereference.

## Where to go next

- New here? Read the [introduction](/getting-started/introduction/), then the
  [quick start](/getting-started/quick-start/).
- Installing? See [installation](/getting-started/installation/).
- Doing a specific job? The [guides](/guides/) are task-first.
- Need every flag? The [CLI reference](/reference/cli/) is the full surface.
