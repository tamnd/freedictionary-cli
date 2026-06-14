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
		name      string
		phonetics []ph
		wantText  string
		wantAudio string
	}{
		{
			name:      "empty slice",
			phonetics: nil,
			wantText:  "",
			wantAudio: "",
		},
		{
			name:      "first has audio",
			phonetics: []ph{{Text: "/hɛloʊ/", Audio: "https://example.com/hello.mp3"}},
			wantText:  "/hɛloʊ/",
			wantAudio: "https://example.com/hello.mp3",
		},
		{
			name: "second has audio",
			phonetics: []ph{
				{Text: "/hɛloʊ/", Audio: ""},
				{Text: "/həˈloʊ/", Audio: "https://example.com/hello-us.mp3"},
			},
			wantText:  "/həˈloʊ/",
			wantAudio: "https://example.com/hello-us.mp3",
		},
		{
			name:      "no audio fallback to first text",
			phonetics: []ph{{Text: "/hɛloʊ/", Audio: ""}, {Text: "/həˈloʊ/", Audio: ""}},
			wantText:  "/hɛloʊ/",
			wantAudio: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Adapt to the unexported struct slice type used by bestPhonetic.
			adapted := make([]struct {
				Text  string `json:"text"`
				Audio string `json:"audio"`
			}, len(tc.phonetics))
			for i, p := range tc.phonetics {
				adapted[i].Text = p.Text
				adapted[i].Audio = p.Audio
			}
			gotText, gotAudio := bestPhonetic(adapted)
			if gotText != tc.wantText {
				t.Errorf("text = %q, want %q", gotText, tc.wantText)
			}
			if gotAudio != tc.wantAudio {
				t.Errorf("audio = %q, want %q", gotAudio, tc.wantAudio)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	got := unique([]string{"b", "a", "b", "", "c", "a"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("unique = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("unique[%d] = %q, want %q", i, got[i], want[i])
		}
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

	defs := toDefinitions(entry, "en")
	if len(defs) != 2 {
		t.Fatalf("len(defs) = %d, want 2", len(defs))
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
	if d.Audio != "https://example.com/test.mp3" {
		t.Errorf("Audio = %q", d.Audio)
	}
	if d.Language != "en" {
		t.Errorf("Language = %q, want en", d.Language)
	}
	if d.SourceURL != "https://en.wiktionary.org/wiki/test" {
		t.Errorf("SourceURL = %q", d.SourceURL)
	}
	// synonyms from meaning + definition merged and sorted
	wantSyns := []string{"exam", "trial"}
	if len(d.Synonyms) != len(wantSyns) {
		t.Errorf("Synonyms = %v, want %v", d.Synonyms, wantSyns)
	}

	d2 := defs[1]
	if d2.PartOfSpeech != "verb" {
		t.Errorf("defs[1].PartOfSpeech = %q, want verb", d2.PartOfSpeech)
	}
}
