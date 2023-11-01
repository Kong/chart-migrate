package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-logr/zapr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/kong/chart-migrate/pkg/core"
)

func main() {
	Execute()
}

// Execute is the entry point to the controller manager.
func Execute() {
	var (
		cfg     core.Config
		rootCmd = GetRootCmd(&cfg)
	)
	cobra.CheckErr(rootCmd.Execute())
}

func GetRootCmd(cfg *core.Config) *cobra.Command {
	cmd := &cobra.Command{
		PersistentPreRunE: bindEnvVars,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd.Context(), cfg, os.Stderr)
		},
		SilenceUsage: true,
		// We can silence the errors because cobra.CheckErr below will print
		// the returned error and set the exit code to 1.
		SilenceErrors: true,
	}
	cmd.Flags().AddFlagSet(cfg.FlagSet())
	return cmd
}

// Run runs the migration application.
func Run(ctx context.Context, c *core.Config, output io.Writer) error {
	// TODO make a logger that doesn't dump stack traces
	logbase, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	logger := zapr.NewLoggerWithOptions(logbase, zapr.LogInfoLevel("v"))
	return core.Run(ctx, c, logger)
}

// == Envvar binding

const envKeyPrefix = "CHART_MIGRATE_"

// bindEnvVars is the simplified viper bind alternative used in KIC.
func bindEnvVars(cmd *cobra.Command, _ []string) (err error) {
	var envKey string
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("environment binding failed for variable %s: %v", envKey, r)
		}
	}()

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		envKey = fmt.Sprintf("%s%s", envKeyPrefix, strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_")))

		if f.Changed {
			return // flags take precedence over environment variables
		}

		if envValue, envSet := os.LookupEnv(envKey); envSet {
			if err := f.Value.Set(envValue); err != nil {
				panic(err)
			}
		}
	})

	return
}
