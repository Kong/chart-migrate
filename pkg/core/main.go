package core

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
)

type Config struct {
	// SourceChart indicates whether the values.yaml to migrate comes from the "kong" or "ingress" chart
	SourceChart string

	flagSet *pflag.FlagSet
}

func (c *Config) FlagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)

	flagSet.StringVar(&c.SourceChart, "source-chart", "kong", `The chart of the original values.yaml, either "kong" or "ingress".`)

	c.flagSet = flagSet
	return flagSet
}

func Run(_ context.Context, _ *Config, _ logr.Logger) error {
	return nil
}
