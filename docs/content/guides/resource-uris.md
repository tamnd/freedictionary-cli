---
title: "Resource URIs"
description: "Use freedictionary as a database/sql-style driver so a host program can address word definitions as freedictionary:// URIs."
weight: 20
---

`freedictionary` is a command line, but the `freedictionary` Go package is also
a small driver that makes word definitions addressable as a resource URI. A host
program registers it the way a program registers a database driver with
`database/sql`, then dereferences `freedictionary://` URIs without knowing
anything about how the API is called.

The host that does this today is [ant](https://github.com/tamnd/ant), a single
binary that puts one URI namespace over a family of site tools. The examples
below use `ant`; any program that links the package gets the same behaviour.

## Mounting the driver

A host enables the driver with one blank import, exactly like `import _
"github.com/lib/pq"`:

```go
import _ "github.com/tamnd/freedictionary-cli/freedictionary"
```

The package's `init` registers a domain with the scheme `freedictionary` for the
host `api.dictionaryapi.dev`. The standalone `freedictionary` binary does not
change.

## Addressing records

A URI is `scheme://authority/id`. The scheme is `freedictionary`:

| URI                                   | What it is                              |
| ------------------------------------- | --------------------------------------- |
| `freedictionary://word/hello`         | all meanings for "hello" (English)      |

```bash
ant get freedictionary://word/hello        # all definitions as records
ant url freedictionary://word/hello        # the dictionary.com browse URL
ant resolve hello                          # bare word, back to its URI
```

## Why this is the same code

The driver and the binary share one definition per operation. A define op
answers both `freedictionary define hello` on the command line and
`ant get freedictionary://word/hello` through a host, from the same handler
and the same client. There is no second implementation to keep in step.
