package main

import (
	"context"
	"flag" //nolint:depguard // We only allow to import the flag package in here
	"fmt"
	"log" //nolint:depguard // TODO: Replace by log/slog
	"os"

	"github.com/leonklingele/sockstun"
	"github.com/leonklingele/sockstun/cmd/sockstun/pathutil"

	"github.com/pkg/errors"
)

const (
	version = "1.0.0-unreleased"

	defaultConfigFilePath = "$HOME/.sockstun/config.toml"
)

func run(cfp string) error {
	cfp, err := annotateConfigFilePath(cfp)
	if err != nil {
		return errors.Wrap(err, "failed to annotate config file path")
	}
	c, err := loadConfig(cfp)
	if err != nil {
		return err
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)
	st, err := sockstun.New(c.SOCKSURI, c.RWTimeout.Duration, logger)
	if err != nil {
		return errors.Wrap(err, "failed to create sockstun")
	}
	for n, r := range c.Rules {
		st.Add(n, r.LocalSock, r.RemoteSock)
	}
	ctx := context.Background()
	return st.Run(ctx) //nolint:wrapcheck // No need to wrap this
}

func main() {
	printVersion := flag.Bool("version", false, "show version and exit")
	configFilePath := flag.String("config", defaultConfigFilePath, "optional, path to config file")
	flag.Parse()

	if *printVersion {
		fmt.Printf("v%s\n", version) //nolint:forbidigo,revive,errcheck // Easiest way to print to stdout
		os.Exit(0)
	}

	if err := run(*configFilePath); err != nil {
		log.Fatal(err)
	}
}

func annotateConfigFilePath(p string) (string, error) {
	return pathutil.ReplaceHome(p) //nolint:wrapcheck // No need to wrap this
}
