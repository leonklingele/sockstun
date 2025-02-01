package main

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

type config struct {
	SOCKSURI  string          `toml:"socks_uri"`
	RWTimeout duration        `toml:"rw_timeout"`
	Rules     map[string]rule `toml:"rules"`
}

func loadConfig(cfp string) (*config, error) {
	var conf config
	if _, err := toml.DecodeFile(cfp, &conf); err != nil {
		return nil, fmt.Errorf("failed to load or parse config file: %w", err)
	}
	return &conf, nil
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}
	return nil
}

//nolint:unparam // Required to satisfy the interface
func (d duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}
