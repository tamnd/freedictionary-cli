---
title: "Quick start"
description: "Look up your first word with freedictionary."
weight: 30
---

Once `freedictionary` is on your `PATH`, look up a word:

```bash
freedictionary define hello
```

By default you get an aligned table with one row per meaning. Ask for JSON when
you want to pipe it:

```bash
$ freedictionary define hello -o json
[
  {
    "word": "hello",
    "phonetic": "/həˈloʊ/",
    "audio": "https://api.dictionaryapi.dev/media/pronunciations/en/hello-us.mp3",
    "part_of_speech": "exclamation",
    "definition": "Used as a greeting or to begin a telephone conversation.",
    "example": "hello there, Katie!",
    "synonyms": ["hi", "howdy"],
    "antonyms": [],
    "language": "en",
    "source_url": "https://en.wiktionary.org/wiki/hello"
  },
  ...
]
```

## Change the language

Use `--lang` to look up a word in another language:

```bash
freedictionary define bonjour --lang fr
freedictionary define hola --lang es
freedictionary define 안녕하세요 --lang ko
```

Supported codes: `en`, `hi`, `es`, `fr`, `ja`, `ru`, `de`, `it`, `ko`,
`pt-BR`, `ar`, `tr`.

## Shape the output

The same flags work on every command:

```bash
freedictionary define hello --fields word,part_of_speech,definition
freedictionary define hello --template '{{.Definition}}'
freedictionary define hello -o jsonl | jq .synonyms
```

`-o` takes `table`, `json`, `jsonl`, `csv`, `tsv`, `url`, or `raw`. Left to
`auto`, it prints a table to a terminal and JSONL into a pipe, so the same
command reads well by hand and parses cleanly downstream.

## Serve it instead

The same operations are available over HTTP and to agents over MCP:

```bash
freedictionary serve --addr :7777 &
curl -s 'localhost:7777/v1/define/hello'      # NDJSON, one record per line
freedictionary mcp                            # MCP over stdio: define tool
```
