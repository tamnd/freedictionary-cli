---
title: "Output formats"
description: "The output contract every command shares: formats, fields, and templates."
weight: 30
---

Every list command renders through one formatter, so the same flags work
everywhere. Pick a format with `-o`, or let `freedictionary` choose: a table
when writing to a terminal, JSONL when piped.

## Formats

```bash
freedictionary define hello -o table   # aligned columns for reading
freedictionary define hello -o jsonl   # one JSON object per line, for piping
freedictionary define hello -o json    # a single JSON array
freedictionary define hello -o csv     # spreadsheet friendly
freedictionary define hello -o tsv     # tab-separated
freedictionary define hello -o url     # just the source_url column
freedictionary define hello -o raw     # the underlying bytes, unformatted
```

| Format | Best for |
|---|---|
| `table` | Reading on a terminal |
| `jsonl` | Piping into another tool, one object at a time |
| `json` | Loading a whole result as an array |
| `csv` / `tsv` | Spreadsheets and quick column math |
| `url` | Feeding URLs into other commands |
| `raw` | The unformatted bytes |

## Narrowing columns

Keep only the fields you want:

```bash
freedictionary define hello --fields word,part_of_speech,definition
```

`--no-header` drops the header row in `table` and `csv` output.

## Templating rows

For full control over each line, apply a Go text/template. Fields are the JSON
keys, capitalised:

```bash
freedictionary define hello --template '{{.PartOfSpeech}}: {{.Definition}}'
```

## Why auto-detection helps

Because the default adapts to the destination, the same command reads well by
hand and parses cleanly in a pipe:

```bash
freedictionary define hello            # a table, because this is a terminal
freedictionary define hello | wc -l    # JSONL, because this is a pipe
```

You only reach for `-o` when you want something other than that default.
