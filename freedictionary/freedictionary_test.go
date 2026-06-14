package freedictionary_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/freedictionary-cli/freedictionary"
)

const fakeDefineJSON = `[
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
          {"definition": "Used as a greeting.", "example": "Hello there!", "synonyms": ["hi"], "antonyms": []}
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

func TestDefineSendsUA(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeDefineJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Define(context.Background(), "hello", "en")
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestDefineURL(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeDefineJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Define(context.Background(), "hello", "en")
	if err != nil {
		t.Fatal(err)
	}
	want := "/api/v2/entries/en/hello"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestDefineURLLang(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeDefineJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Define(context.Background(), "bonjour", "fr")
	if err != nil {
		t.Fatal(err)
	}
	want := "/api/v2/entries/fr/bonjour"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestDefineParsesMeanings(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeDefineJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	defs, err := c.Define(context.Background(), "hello", "en")
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) != 2 {
		t.Fatalf("len(defs) = %d, want 2", len(defs))
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
	if d.Audio == "" {
		t.Error("Audio is empty")
	}
	if d.Language != "en" {
		t.Errorf("Language = %q, want en", d.Language)
	}
	if d.SourceURL != "https://en.wiktionary.org/wiki/hello" {
		t.Errorf("SourceURL = %q", d.SourceURL)
	}
}

func TestDefineNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, fakeNotFoundJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Define(context.Background(), "xyznotaword", "en")
	if err == nil {
		t.Fatal("expected error for not-found word, got nil")
	}
}

func TestDefineRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeDefineJSON)
	}))
	defer ts.Close()

	cfg := freedictionary.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := freedictionary.NewClient(cfg)

	_, err := c.Define(context.Background(), "hello", "en")
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

func TestDefineSynonymsDeduped(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeDefineJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	defs, err := c.Define(context.Background(), "hello", "en")
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) == 0 {
		t.Fatal("no definitions returned")
	}
	// Check that synonyms from meaning-level and definition-level are merged.
	d := defs[0]
	synMap := make(map[string]bool)
	for _, s := range d.Synonyms {
		if synMap[s] {
			t.Errorf("duplicate synonym %q in %v", s, d.Synonyms)
		}
		synMap[s] = true
	}
}
