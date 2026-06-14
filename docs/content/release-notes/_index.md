---
title: "Release notes"
linkTitle: "Release notes"
description: "What changed in each freedictionary release, newest first."
weight: 40
---

What shipped in each release, newest first. Every tagged version builds the same
set of artifacts: archives for Linux, macOS, Windows, and FreeBSD, Linux
packages (deb, rpm, apk), a multi-arch container image on GHCR, and entries for
the package managers. Binaries are pure Go, so there is nothing to install
alongside them.

## v0.1.0

First release. Ships one command:

- `define <word> [--lang <code>]` — look up a word's definitions from the
  Free Dictionary API (api.dictionaryapi.dev). Returns one record per meaning,
  with the word, phonetic, audio URL, part of speech, definition, example,
  synonyms, antonyms, language, and source URL.
