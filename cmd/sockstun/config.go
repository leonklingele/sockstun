package main

import (
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type config struct {
	SOCKSURI  string          `toml:"socks_uri"`
	RWTimeout duration        `toml:"rw_timeout"`
	Rules     map[string]rule `toml:"rules"`
}

func loadConfig(cfp string) (*config, error) {
	var conf config
	if _, err := toml.DecodeFile(cfp, &conf); err != nil {
		return nil, errors.Wrap(err, "failed to load or parse config file")
	}
	return &conf, nil
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

//nolint:unparam
func (d duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}
