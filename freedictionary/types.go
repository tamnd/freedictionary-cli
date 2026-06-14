// Package freedictionary is the library behind the freedictionary command line:
// the HTTP client, request shaping, and the typed data models for the Free
// Dictionary API (api.dictionaryapi.dev).
//
// The API requires no authentication. A polite User-Agent and 100 ms pacing
// between requests keeps the client within the free-tier rate limits.
package freedictionary

// Definition is one record produced by the Lookup call.
// Each entry in the API response, times each meaning, times each definition
// yields one Definition record.
type Definition struct {
	Word         string `kit:"id" json:"word"`
	PartOfSpeech string `json:"part_of_speech"`
	Definition   string `json:"definition"`
	Example      string `json:"example"`
	Synonyms     string `json:"synonyms"` // comma-joined, up to 5
	Phonetic     string `json:"phonetic"`
}

// wireEntry is the wire shape returned by the Free Dictionary API for one word
// entry. The API returns an array of these.
type wireEntry struct {
	Word      string `json:"word"`
	Phonetic  string `json:"phonetic"`
	Phonetics []struct {
		Text  string `json:"text"`
		Audio string `json:"audio"`
	} `json:"phonetics"`
	Meanings []struct {
		PartOfSpeech string `json:"partOfSpeech"`
		Definitions  []struct {
			Definition string   `json:"definition"`
			Example    string   `json:"example"`
			Synonyms   []string `json:"synonyms"`
			Antonyms   []string `json:"antonyms"`
		} `json:"definitions"`
		Synonyms []string `json:"synonyms"`
		Antonyms []string `json:"antonyms"`
	} `json:"meanings"`
	SourceUrls []string `json:"sourceUrls"`
}

// wireError is the shape returned by the API on 404.
type wireError struct {
	Title      string `json:"title"`
	Message    string `json:"message"`
	Resolution string `json:"resolution"`
}
