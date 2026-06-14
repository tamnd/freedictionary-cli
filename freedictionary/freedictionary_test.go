package freedictionary_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/freedictionary-cli/freedictionary"
)

const fakeLookupJSON = `[
  {
    "word": "hello",
    "phonetic": "/hɛloʊ/",
    "phonetics": [
      {"text": "/həˈloʊ/", "audio": "https://api.dictionaryapi.dev/media/pronunciations/en/hello-us.mp3"},
      {"text": "/həˈləʊ/", "audio": "https://api.dictionaryapi.dev/media/pronunciations/en/hello-uk.mp3"}
    ],
    "meanings": [
      {
        "partOfSpeech": "exclamation",
        "definitions": [
          {"definition": "Used as a greeting.", "example": "Hello there!", "synonyms": ["hi"], "antonyms": []},
          {"definition": "Used to attract attention.", "example": "Hello? Anyone there?", "synonyms": [], "antonyms": []}
        ],
        "synonyms": ["howdy"],
        "antonyms": []
      },
      {
        "partOfSpeech": "noun",
        "definitions": [
          {"definition": "An utterance of hello.", "example": "", "synonyms": [], "antonyms": []}
        ],
        "synonyms": [],
        "antonyms": []
      }
    ],
    "sourceUrls": ["https://en.wiktionary.org/wiki/hello"]
  }
]`

const fakeNotFoundJSON = `{"title":"No Definitions Found","message":"Sorry pal, we couldn't find definitions for the word you were looking for.","resolution":"You can try the search again at later time or head to the web instead."}`

func newTestClient(ts *httptest.Server) *freedictionary.Client {
	cfg := freedictionary.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return freedictionary.NewClient(cfg)
}

func TestLookupSendsUA(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeLookupJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Lookup(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestLookupURL(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeLookupJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Lookup(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	want := "/api/v2/entries/en/hello"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

// TestLookupExpandsDefinitions checks that one record is emitted per
// (meaning x definition) pair, not one per meaning.
func TestLookupExpandsDefinitions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeLookupJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	defs, err := c.Lookup(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	// exclamation has 2 defs, noun has 1 def -> total 3
	if len(defs) != 3 {
		t.Fatalf("len(defs) = %d, want 3", len(defs))
	}

	d := defs[0]
	if d.Word != "hello" {
		t.Errorf("Word = %q, want hello", d.Word)
	}
	if d.PartOfSpeech != "exclamation" {
		t.Errorf("PartOfSpeech = %q, want exclamation", d.PartOfSpeech)
	}
	if d.Definition != "Used as a greeting." {
		t.Errorf("Definition = %q", d.Definition)
	}
	if d.Example != "Hello there!" {
		t.Errorf("Example = %q", d.Example)
	}
	if d.Phonetic == "" {
		t.Error("Phonetic should not be empty")
	}
}

func TestLookupSynonymsString(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeLookupJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	defs, err := c.Lookup(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	// defs[0] should have synonyms from meaning ("howdy") + definition ("hi")
	d := defs[0]
	if d.Synonyms == "" {
		t.Error("Synonyms should not be empty for first definition")
	}
	// Must be a comma-joined string, not JSON array notation
	if strings.HasPrefix(d.Synonyms, "[") {
		t.Errorf("Synonyms looks like JSON array: %q", d.Synonyms)
	}
}

func TestLookupNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, fakeNotFoundJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Lookup(context.Background(), "xyzzy")
	if err == nil {
		t.Fatal("expected error for not-found word, got nil")
	}
	if !strings.Contains(err.Error(), "no definitions found") {
		t.Errorf("error message = %q, want it to contain 'no definitions found'", err.Error())
	}
}

func TestLookupRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeLookupJSON)
	}))
	defer ts.Close()

	cfg := freedictionary.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := freedictionary.NewClient(cfg)

	_, err := c.Lookup(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}
