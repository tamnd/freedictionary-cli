---
title: "CLI"
description: "Every command and subcommand, with the flags that matter."
weight: 10
---

```
freedictionary <command> [arguments] [flags]
```

Run `freedictionary <command> --help` for the full flag list on any command.

## Commands

| Command | What it does |
|---|---|
| `define <word>` | Look up a word's definitions (one record per meaning) |
| `serve [--addr]` | Serve the operations over HTTP as NDJSON |
| `mcp` | Run as an MCP server over stdio |
| `version` | Print the version and exit |

## The define command

```
freedictionary define <word> [--lang <code>]
```

Fetches definitions from `api.dictionaryapi.dev` and emits one `Definition`
record per meaning. With three meanings, you get three rows.

| Flag | Default | Meaning |
|---|---|---|
| `--lang` | `en` | Language code: `en`, `hi`, `es`, `fr`, `ja`, `ru`, `de`, `it`, `ko`, `pt-BR`, `ar`, `tr` |

## Global flags

These are shared by every operation.

| Flag | Meaning |
|---|---|
| `-o, --output` | Output format: `auto`, `table`, `json`, `jsonl`, `csv`, `tsv`, `url`, `raw` |
| `--fields` | Comma-separated columns to keep |
| `--template` | Go text/template applied per record |
| `--no-header` | Omit the header row in `table` and `csv` |
| `-n, --limit` | Stop after N records (0 means no limit) |
| `--rate` | Minimum delay between requests |
| `--retries` | Retry attempts on rate limit or 5xx |
| `--timeout` | Per-request timeout |
| `--db` | Tee every record into a store (e.g. `out.db`, `postgres://...`) |
| `-v, --verbose` | Increase verbosity (repeatable) |
| `-q, --quiet` | Suppress progress output |
| `--color` | `auto`, `always`, or `never` |

See [output formats](/reference/output/) for what `-o`, `--fields`, and
`--template` produce.
