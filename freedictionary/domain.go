package freedictionary

import (
	"context"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes the Free Dictionary API as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/freedictionary-cli/freedictionary"
//
// The same Domain also builds the standalone freedictionary binary.
func init() { kit.Register(Domain{}) }

// Domain is the Free Dictionary driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "freedictionary",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "freedictionary",
			Short:  "Look up word definitions from the Free Dictionary API",
			Long: `freedictionary looks up English word definitions from api.dictionaryapi.dev.
No API key required.`,
			Site: Host,
			Repo: "https://github.com/tamnd/freedictionary-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "word",
		Group:   "read",
		List:    true,
		Summary: "Look up definitions for a word",
		Args:    []kit.Arg{{Name: "word", Help: "word to define"}},
	}, wordOp)
}

// newClient builds the client from host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type wordInput struct {
	Word   string  `kit:"arg" help:"word to define"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func wordOp(ctx context.Context, in wordInput, emit func(Definition) error) error {
	defs, err := in.Client.Lookup(ctx, in.Word)
	if err != nil {
		return err
	}
	for _, d := range defs {
		if err := emit(d); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver ---

// Classify turns any input into the canonical (type, id). Every non-empty
// string is treated as a word to look up.
func (Domain) Classify(input string) (uriType, id string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", errs.Usage("empty freedictionary reference")
	}
	return "word", input, nil
}

// Locate returns a reference URL for a word.
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "word":
		return "https://www.dictionary.com/browse/" + id, nil
	default:
		return "", errs.Usage("freedictionary has no resource type %q", uriType)
	}
}
