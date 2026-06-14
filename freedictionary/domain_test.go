package freedictionary

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string functions.
// HTTP behaviour is covered in freedictionary_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "freedictionary" {
		t.Errorf("Scheme = %q, want freedictionary", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "freedictionary" {
		t.Errorf("Identity.Binary = %q, want freedictionary", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}

	typ, id, err := Domain{}.Classify("hello")
	if err != nil || typ != "word" || id != "hello" {
		t.Errorf("Classify = (%q, %q, %v), want (word, hello, nil)", typ, id, err)
	}

	// whitespace trimmed
	typ, id, err = Domain{}.Classify("  world  ")
	if err != nil || typ != "word" || id != "world" {
		t.Errorf("Classify(padded) = (%q, %q, %v), want (word, world, nil)", typ, id, err)
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("word", "hello")
	want := "https://www.dictionary.com/browse/hello"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("unknown", "foo")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}

func TestBestPhonetic(t *testing.T) {
	type ph struct {
		Text  string `json:"text"`
		Audio string `json:"audio"`
	}

	cases := []struct {
		name     string
		input    []ph
		wantText string
	}{
		{name: "empty slice", input: nil, wantText: ""},
		{
			name:     "first has audio",
			input:    []ph{{Text: "/hɛloʊ/", Audio: "https://example.com/hello.mp3"}},
			wantText: "/hɛloʊ/",
		},
		{
			name: "second has audio",
			input: []ph{
				{Text: "/hɛloʊ/", Audio: ""},
				{Text: "/həˈloʊ/", Audio: "https://example.com/hello-us.mp3"},
			},
			wantText: "/həˈloʊ/",
		},
		{
			name:     "no audio fallback to first text",
			input:    []ph{{Text: "/hɛloʊ/", Audio: ""}, {Text: "/həˈloʊ/", Audio: ""}},
			wantText: "/hɛloʊ/",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapted := make([]struct {
				Text  string `json:"text"`
				Audio string `json:"audio"`
			}, len(tc.input))
			for i, p := range tc.input {
				adapted[i].Text = p.Text
				adapted[i].Audio = p.Audio
			}
			got := bestPhonetic(adapted)
			if got != tc.wantText {
				t.Errorf("text = %q, want %q", got, tc.wantText)
			}
		})
	}
}

func TestJoinSynonyms(t *testing.T) {
	// deduplication
	got := joinSynonyms([]string{"hi", "hey"}, []string{"hey", "howdy"}, 5)
	if got != "hi, hey, howdy" {
		t.Errorf("joinSynonyms = %q, want %q", got, "hi, hey, howdy")
	}

	// max cap
	got = joinSynonyms([]string{"a", "b", "c"}, []string{"d", "e", "f"}, 4)
	if got != "a, b, c, d" {
		t.Errorf("joinSynonyms max = %q, want %q", got, "a, b, c, d")
	}

	// empty inputs
	got = joinSynonyms(nil, nil, 5)
	if got != "" {
		t.Errorf("joinSynonyms empty = %q, want empty", got)
	}
}

func TestToDefinitions(t *testing.T) {
	entry := wireEntry{
		Word:     "test",
		Phonetic: "/tɛst/",
		Phonetics: []struct {
			Text  string `json:"text"`
			Audio string `json:"audio"`
		}{
			{Text: "/tɛst/", Audio: "https://example.com/test.mp3"},
		},
		Meanings: []struct {
			PartOfSpeech string `json:"partOfSpeech"`
			Definitions  []struct {
				Definition string   `json:"definition"`
				Example    string   `json:"example"`
				Synonyms   []string `json:"synonyms"`
				Antonyms   []string `json:"antonyms"`
			} `json:"definitions"`
			Synonyms []string `json:"synonyms"`
			Antonyms []string `json:"antonyms"`
		}{
			{
				PartOfSpeech: "noun",
				Definitions: []struct {
					Definition string   `json:"definition"`
					Example    string   `json:"example"`
					Synonyms   []string `json:"synonyms"`
					Antonyms   []string `json:"antonyms"`
				}{
					{Definition: "a procedure for assessment", Example: "take a test", Synonyms: []string{"exam"}, Antonyms: []string{}},
					{Definition: "a short examination", Example: "pass the test", Synonyms: []string{"quiz"}, Antonyms: []string{}},
				},
				Synonyms: []string{"trial"},
				Antonyms: []string{},
			},
			{
				PartOfSpeech: "verb",
				Definitions: []struct {
					Definition string   `json:"definition"`
					Example    string   `json:"example"`
					Synonyms   []string `json:"synonyms"`
					Antonyms   []string `json:"antonyms"`
				}{
					{Definition: "to take a test", Example: "", Synonyms: []string{}, Antonyms: []string{}},
				},
				Synonyms: []string{},
				Antonyms: []string{},
			},
		},
		SourceUrls: []string{"https://en.wiktionary.org/wiki/test"},
	}

	// 2 noun defs + 1 verb def = 3 records
	defs := toDefinitions(entry)
	if len(defs) != 3 {
		t.Fatalf("len(defs) = %d, want 3", len(defs))
	}

	d := defs[0]
	if d.Word != "test" {
		t.Errorf("Word = %q, want test", d.Word)
	}
	if d.PartOfSpeech != "noun" {
		t.Errorf("PartOfSpeech = %q, want noun", d.PartOfSpeech)
	}
	if d.Definition != "a procedure for assessment" {
		t.Errorf("Definition = %q", d.Definition)
	}
	if d.Example != "take a test" {
		t.Errorf("Example = %q", d.Example)
	}
	// meaning synonym "trial" + definition synonym "exam" merged
	if d.Synonyms == "" {
		t.Error("Synonyms should not be empty")
	}

	d2 := defs[2]
	if d2.PartOfSpeech != "verb" {
		t.Errorf("defs[2].PartOfSpeech = %q, want verb", d2.PartOfSpeech)
	}
}
