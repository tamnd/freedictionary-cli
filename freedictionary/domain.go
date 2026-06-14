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
			Long: `freedictionary looks up word definitions from api.dictionaryapi.dev.
No API key required. Supports English and 11 other languages.`,
			Site: Host,
			Repo: "https://github.com/tamnd/freedictionary-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "define",
		Group:   "read",
		List:    true,
		Summary: "Look up a word's definitions",
		Args:    []kit.Arg{{Name: "word", Help: "word to define"}},
	}, defineOp)
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

type defineInput struct {
	Word   string  `kit:"arg" help:"word to define"`
	Lang   string  `kit:"flag" help:"language code (en, es, fr, de, it, ru, ar, hi, ja, ko, pt-BR, tr)" default:"en"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func defineOp(ctx context.Context, in defineInput, emit func(Definition) error) error {
	lang := in.Lang
	if lang == "" {
		lang = "en"
	}
	defs, err := in.Client.Define(ctx, in.Word, lang)
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

// Locate returns the canonical dictionary.com URL for a word.
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "word":
		return "https://www.dictionary.com/browse/" + id, nil
	default:
		return "", errs.Usage("freedictionary has no resource type %q", uriType)
	}
}
