---
title: "Troubleshooting"
description: "The handful of things that trip people up, and how to fix each one."
weight: 40
---

Most of these come down to network reality or how the Free Dictionary API
responds, not a bug.

## No definitions found

The API returns a 404 when it has no entry for the word. Check the spelling,
try a different form (e.g. the base form of a verb), or confirm the word exists
in the language you specified with `--lang`.

## Requests start failing or returning 429

The Free Dictionary API rate-limits under heavy use. `freedictionary` already
paces requests and retries transient failures, but if you are running many
lookups in quick succession you can raise the delay with `--rate` (for example
`--rate 500ms`) and increase retries with `--retries 5`.

## Wrong language results

Make sure you pass `--lang` with the correct code. The default is `en`. For
example, looking up a French word without `--lang fr` will query the English
dictionary and may return no results or an unexpected entry.

## The binary is not on your PATH

`go install` puts the binary in `$(go env GOPATH)/bin` (usually `~/go/bin`), and
a release archive leaves it wherever you unpacked it. If your shell cannot find
`freedictionary`, add that directory to your `PATH`. See
[installation](/getting-started/installation/).

## Seeing what freedictionary actually did

When something behaves unexpectedly, `-v` adds per-request detail so you can
see the URLs it hit and the responses it got. That is usually enough to tell a
rate limit apart from a genuinely empty result.
