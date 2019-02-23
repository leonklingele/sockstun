package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/leonklingele/sockstun"
	"github.com/leonklingele/sockstun/cmd/sockstun/pathutil"
	"github.com/pkg/errors"
)

const (
	version = "1.0.0"

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

	log := log.New(os.Stderr, "", log.LstdFlags)
	st, err := sockstun.New(c.SOCKSURI, c.RWTimeout.Duration, log)
	if err != nil {
		return errors.Wrap(err, "failed to create sockstun")
	}
	for n, r := range c.Rules {
		st.Add(n, r.LocalSock, r.RemoteSock)
	}
	ctx := context.Background()
	return st.Run(ctx)
}

func main() {
	printVersion := flag.Bool("version", false, "show version and exit")
	configFilePath := flag.String("config", defaultConfigFilePath, "optional, path to config file")
	flag.Parse()

	if *printVersion {
		fmt.Printf("v%s\n", version)
		os.Exit(0)
	}

	if err := run(*configFilePath); err != nil {
		log.Fatal(err)
	}
}

func annotateConfigFilePath(p string) (string, error) {
	return pathutil.ReplaceHome(p)
}
