package freedictionary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
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

// Define looks up a word in the given language and returns one Definition per
// meaning. On 404 the API returns a JSON error body; this returns a descriptive
// error message instead of a raw HTTP error.
func (c *Client) Define(ctx context.Context, word, lang string) ([]Definition, error) {
	u := fmt.Sprintf("%s/api/v2/entries/%s/%s", c.cfg.BaseURL, lang, word)
	body, status, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		var we wireError
		if jerr := json.Unmarshal(body, &we); jerr == nil && we.Title != "" {
			return nil, fmt.Errorf("%s", we.Title)
		}
		return nil, fmt.Errorf("no definitions found for %q", word)
	}
	var entries []wireEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	var defs []Definition
	for _, e := range entries {
		defs = append(defs, toDefinitions(e, lang)...)
	}
	return defs, nil
}

// toDefinitions converts one wireEntry into one Definition per meaning.
func toDefinitions(entry wireEntry, lang string) []Definition {
	phText, phAudio := bestPhonetic(entry.Phonetics)
	if phText == "" {
		phText = entry.Phonetic
	}
	source := ""
	if len(entry.SourceUrls) > 0 {
		source = entry.SourceUrls[0]
	}
	var defs []Definition
	for _, m := range entry.Meanings {
		syns := make([]string, len(m.Synonyms))
		copy(syns, m.Synonyms)
		ants := make([]string, len(m.Antonyms))
		copy(ants, m.Antonyms)
		def := ""
		ex := ""
		if len(m.Definitions) > 0 {
			def = m.Definitions[0].Definition
			ex = m.Definitions[0].Example
			syns = append(syns, m.Definitions[0].Synonyms...)
			ants = append(ants, m.Definitions[0].Antonyms...)
		}
		defs = append(defs, Definition{
			Word:         entry.Word,
			Phonetic:     phText,
			Audio:        phAudio,
			PartOfSpeech: m.PartOfSpeech,
			Definition:   def,
			Example:      ex,
			Synonyms:     unique(syns),
			Antonyms:     unique(ants),
			Language:     lang,
			SourceURL:    source,
		})
	}
	return defs
}

// bestPhonetic picks the phonetic text and audio URL from a phonetics slice.
// It prefers the first entry that has a non-empty audio URL. If none has audio,
// it falls back to the text of the first entry.
func bestPhonetic(phonetics []struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}) (text, audio string) {
	for _, p := range phonetics {
		if p.Audio != "" && text == "" {
			text = p.Text
			audio = p.Audio
		}
	}
	if text == "" && len(phonetics) > 0 {
		text = phonetics[0].Text
	}
	return
}

// unique returns s deduplicated and sorted, with empty strings removed.
func unique(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	out := s[:0:0]
	for _, v := range s {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
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
	req.Header.Set("User-Agent", "freedictionary-cli/0.1 (github.com/tamnd/freedictionary-cli)")
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
