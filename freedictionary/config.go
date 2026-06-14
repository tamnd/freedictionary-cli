package freedictionary

import "time"

// Host is the site this client talks to.
const Host = "api.dictionaryapi.dev"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL string
	Rate    time.Duration
	Retries int
	Timeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL: "https://api.dictionaryapi.dev",
		Rate:    100 * time.Millisecond,
		Retries: 3,
		Timeout: 15 * time.Second,
	}
}
