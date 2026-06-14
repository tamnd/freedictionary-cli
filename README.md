# freedictionary

A command line for the [Free Dictionary API](https://dictionaryapi.dev/). No API key required. Supports English and 11 other languages.

```bash
freedictionary define hello
freedictionary define hello -o json
freedictionary define bonjour --lang fr
freedictionary define hello -o jsonl | jq .synonyms
```

Built on [any-cli/kit](https://github.com/tamnd/any-cli): one operation is a CLI command, an HTTP route, an MCP tool, and a resource-URI dereference.

## Install

```bash
go install github.com/tamnd/freedictionary-cli/cmd/freedictionary@latest
```

Or download a prebuilt binary from the [releases page](https://github.com/tamnd/freedictionary-cli/releases).

## Usage

```
freedictionary define <word> [--lang <code>]
```

Returns one record per meaning. Each record has: word, phonetic, audio URL, part of speech, definition, example, synonyms, antonyms, language, source URL.

Supported language codes: `en` (default), `hi`, `es`, `fr`, `ja`, `ru`, `de`, `it`, `ko`, `pt-BR`, `ar`, `tr`.

## Output

```bash
freedictionary define hello                          # table on terminal, JSONL in pipe
freedictionary define hello -o json                  # JSON array
freedictionary define hello --fields word,definition # keep two columns
freedictionary define hello --template '{{.PartOfSpeech}}: {{.Definition}}'
```

## As an HTTP server or MCP tool

```bash
freedictionary serve --addr :7777
curl -s 'localhost:7777/v1/define/hello'

freedictionary mcp    # MCP over stdio
```

## As a resource-URI driver

```go
import _ "github.com/tamnd/freedictionary-cli/freedictionary"
```

Registers the `freedictionary` scheme so a host (like [ant](https://github.com/tamnd/ant)) can address words as `freedictionary://word/hello`.

## License

Apache 2.0
