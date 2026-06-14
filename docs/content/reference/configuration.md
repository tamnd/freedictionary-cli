---
title: "Configuration"
description: "Environment variables, defaults, and the data directory."
weight: 20
---

`freedictionary` needs almost no configuration: it runs anonymously against
the public Free Dictionary API out of the box. The settings below let you tune
politeness and storage.

## Defaults

| Setting | Default | Flag |
|---|---|---|
| Language | `en` | `--lang` |
| Requests | paced at 100 ms, retried on 429/5xx | `--rate`, `--retries` |
| Per-request timeout | 15s | `--timeout` |
| Retry attempts | 3 | `--retries` |

## Environment variables

Every flag has an environment fallback, prefixed `FREEDICTIONARY_` in
upper case with dashes as underscores. For example:

```bash
export FREEDICTIONARY_RATE=500ms
export FREEDICTIONARY_TIMEOUT=30s
```

Flags win over environment variables, which win over the built-in defaults.

## Sending records to a store

`--db` tees every emitted record into a store as a side effect of reading, so a
session fills a local database without a separate import step:

```bash
freedictionary define hello --db out.db          # SQLite file
freedictionary define hello --db 'postgres://...'
```
