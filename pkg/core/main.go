package core

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"sigs.k8s.io/yaml"
)

// NOTE some limitations of this approach:
// - comments go poof because of the YAML->JSON->YAML conversion. Don't think we can save those without a pure YAML
//   workflow. Hopefully matters less for user values.yamls, but I expect at least some have their own comments that
//   would have to be re-added manually.
// - existing and new keys are simply alphabetized at some point. preserving those would presumably require keeping
//   track of indices for each and finding some manipulation tool that does preserve them. structs probably do this
//   but we're not operating on structs.

type Config struct {
	// SourceChart indicates whether the values.yaml to migrate comes from the "kong" or "ingress" chart
	SourceChart string

	// InputFile is the values.yaml filename to migrate.
	InputFile string

	// OutputFormat is the output format.
	OutputFormat string

	flagSet *pflag.FlagSet
}

func (c *Config) FlagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)

	flagSet.StringVarP(&c.SourceChart, "source-chart", "s", "kong", `The chart of the original values.yaml, either "kong" or "ingress".`)
	flagSet.StringVarP(&c.InputFile, "file", "f", "./values.yaml", `Path to the values.yaml to transform.`)
	flagSet.StringVar(&c.OutputFormat, "output-format", "yaml", `Output format, either "yaml" (default) or "json"`)

	c.flagSet = flagSet
	return flagSet
}

const (
	kongChart    = "kong"
	ingressChart = "ingress"
)

// RunOut runs Run and prints its result to stdout.
func RunOut(ctx context.Context, c *Config, logger logr.Logger) error {
	output, err := Run(ctx, c, logger)
	if err != nil {
		return err
	}
	fmt.Printf("%s", string(output))
	return nil

}

// Run takes a configuration string and returns a migrated configuration string.
func Run(_ context.Context, c *Config, logger logr.Logger) (string, error) {
	raw, err := os.ReadFile(c.InputFile)
	if err != nil {
		return "", fmt.Errorf("could not read %s: %w", c.InputFile, err)
	}
	orig, err := yaml.YAMLToJSON(raw)
	if err != nil {
		return "", fmt.Errorf("could not parse input values.yaml YAML into JSON: %w", err)
	}

	// Keep a copy of the original to diff later.
	transformed := orig

	var remaps mapFunc
	switch originChart := c.SourceChart; originChart {
	case ingressChart:
		remaps = getIngressKeys
	case kongChart:
		remaps = getKongKeys
	default:
		return "", fmt.Errorf("unknown source chart: %s", originChart)
	}

	for start, end := range remaps() {
		var found bool
		transformed, found, err = Move(start, end, transformed)
		if err != nil {
			logger.Error(err, "migration failed")
		}
		if found {
			// not immediately clear why, but attempting to move AND delete within Move (the contents of Delete originally
			// followed the sjson.SetBytes() call and error check) resulted in it deleting both the old and new key.
			// Presumably something about how it addresses the values internally. Returning and then deleting avoids this,
			// since we have a new []byte to work with.
			transformed, err = Delete(start, transformed)
			if err != nil {
				logger.Error(err, "cleanup failed")
			}
		}
	}

	if c.OutputFormat == "json" {
		return string(transformed), nil
	}
	yamlOut, err := yaml.JSONToYAML(transformed)
	if err != nil {
		return "", fmt.Errorf("could not convert back to YAML: %w", err)
	}
	return string(yamlOut), nil
}

// Move takes a start and end JSON path string and a document, and returns a document with the start path moved to the
// end path. It also returns a boolean indicating if the start path is not present, in which case the input document is
// returned unmodified.
func Move(start, end string, doc []byte) ([]byte, bool, error) {
	found := gjson.GetBytes(doc, start)
	if !found.Exists() {
		return doc, false, nil
	}

	result, err := sjson.SetBytes(doc, end, found.Value())
	if err != nil {
		return doc, true, fmt.Errorf("could not inject %s at %s: %w", start, end, err)
	}

	return result, true, nil
}

// Delete takes a JSON path and JSON document and returns a JSON document with the input path deleted.
func Delete(key string, doc []byte) ([]byte, error) {
	result, err := sjson.DeleteBytes(doc, key)
	if err != nil {
		return doc, fmt.Errorf("could not delete old content at %s: %w", key, err)
	}
	return result, nil
}
