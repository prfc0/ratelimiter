package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Backend BackendConfig `json:"backend"`
	Rules   []Rule        `json:"rules"`
}

type BackendConfig struct {
	URL string `json:"url"`
}

type Rule struct {
	Name       string `json:"name"`
	Algorithm  string `json:"algorithm"`
	KeySource  string `json:"key_source"`            // "ip" | "header" | "path"
	HeaderName string `json:"header_name,omitempty"` // used if key_source == "header"`

	// Common params
	Limit         int `json:"limit,omitempty"`
	WindowSeconds int `json:"window_seconds,omitempty"`

	// Token/Leaky bucket
	Capacity   int `json:"capacity,omitempty"`
	RefillRate int `json:"refill_rate,omitempty"`
	LeakRate   int `json:"leak_rate,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
