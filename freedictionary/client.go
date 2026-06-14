package freedictionary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client talks to the Free Dictionary API over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Lookup looks up a word and returns one Definition per (meaning x definition)
// combination. On 404 it returns an error with the message
// "no definitions found for %q".
func (c *Client) Lookup(ctx context.Context, word string) ([]Definition, error) {
	u := fmt.Sprintf("%s/api/v2/entries/en/%s", c.cfg.BaseURL, word)
	body, status, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		var we wireError
		if jerr := json.Unmarshal(body, &we); jerr == nil && we.Title != "" {
			return nil, fmt.Errorf("no definitions found for %q", word)
		}
		return nil, fmt.Errorf("no definitions found for %q", word)
	}
	var entries []wireEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	var defs []Definition
	for _, e := range entries {
		defs = append(defs, toDefinitions(e)...)
	}
	return defs, nil
}

// toDefinitions converts one wireEntry into one Definition per
// (meaning, definition) pair.
func toDefinitions(entry wireEntry) []Definition {
	phonetic := bestPhonetic(entry.Phonetics)
	if phonetic == "" {
		phonetic = entry.Phonetic
	}
	var defs []Definition
	for _, m := range entry.Meanings {
		for _, d := range m.Definitions {
			// Collect synonyms from meaning-level and definition-level, deduplicated, max 5.
			syns := joinSynonyms(m.Synonyms, d.Synonyms, 5)
			defs = append(defs, Definition{
				Word:         entry.Word,
				Phonetic:     phonetic,
				PartOfSpeech: m.PartOfSpeech,
				Definition:   d.Definition,
				Example:      d.Example,
				Synonyms:     syns,
			})
		}
	}
	return defs
}

// bestPhonetic picks the phonetic text from a phonetics slice.
// It prefers the first entry that has a non-empty audio URL; falls back to
// the text of the first entry with any text.
func bestPhonetic(phonetics []struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}) string {
	for _, p := range phonetics {
		if p.Audio != "" && p.Text != "" {
			return p.Text
		}
	}
	for _, p := range phonetics {
		if p.Text != "" {
			return p.Text
		}
	}
	return ""
}

// joinSynonyms merges two synonym slices, deduplicates, and returns the first
// max entries as a comma-joined string.
func joinSynonyms(a, b []string, max int) string {
	seen := make(map[string]struct{}, len(a)+len(b))
	var out []string
	for _, s := range append(a, b...) {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
		if len(out) >= max {
			break
		}
	}
	return strings.Join(out, ", ")
}

// get performs a retrying GET and returns (body, status, error).
func (c *Client) get(ctx context.Context, url string) ([]byte, int, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, status, retry, err := c.do(ctx, url)
		if err == nil {
			return body, status, nil
		}
		lastErr = err
		if !retry {
			return nil, status, err
		}
	}
	return nil, 0, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, int, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, false, err
	}
	req.Header.Set("User-Agent", "freedictionary-cli/0.1 (tamnd87@gmail.com)")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, true, err
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, resp.StatusCode, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	// 404 is a valid domain response (word not found), not a retry-able error.
	if resp.StatusCode == http.StatusNotFound {
		return b, resp.StatusCode, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	return b, resp.StatusCode, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		return 5 * time.Second
	}
	return d
}
