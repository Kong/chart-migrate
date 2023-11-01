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

// TODO a proper type would be nice, but I don't think we can create a perfect one.
// aside from annoyances where some keys are commented by default,
// values.yaml is not a regular struct. Some sections like .env allow for
// arbitrary keys. We can presumably map[string]interface{} these, however.
// Unsure if there's a benefit to a struct beyond just map[string]interface{}
// everything and assuming found keys contain valid values. This is reasonable,
// since invalid values wouldn't have rendered in Helm.
//
// the gjson/sjson approach below may well be better for a limited set of keys,
// since most everything we want to just leave in place, and it does that

type Config struct {
	// SourceChart indicates whether the values.yaml to migrate comes from the "kong" or "ingress" chart
	SourceChart string

	// InputFile is the values.yaml filename to migrate.
	InputFile string

	flagSet *pflag.FlagSet
}

func (c *Config) FlagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)

	flagSet.StringVarP(&c.SourceChart, "source-chart", "s", "kong", `The chart of the original values.yaml, either "kong" or "ingress".`)
	flagSet.StringVarP(&c.InputFile, "file", "f", "./values.yaml", `Path to the values.yaml to transform.`)

	c.flagSet = flagSet
	return flagSet
}

// getKeyReMaps returns a map of strings to strings. Keys are the original locations of a key in values.yaml and values
// are their new locations. Both are in dotted string format: "foo.bar.baz" indicates a YAML structure like:
// foo:
//
//	bar:
//	  baz: {}
func getKeyReMaps() map[string]string {
	return map[string]string{
		"podAnnotations":        "deployment.kong.pod.annotations",
		"deploymentAnnotations": "deployment.kong.annotations",
	}
}

func Run(_ context.Context, c *Config, logger logr.Logger) error {
	input, err := os.Open(c.InputFile)
	if err != nil {
		return fmt.Errorf("could not open input values.yaml: %w", err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return fmt.Errorf("could not inspect input values.yaml: %w", err)
	}

	raw := make([]byte, info.Size())
	var orig, transformed []byte
	_, err = input.Read(raw)
	if err != nil {
		return fmt.Errorf("could not read input values.yaml: %w", err)
	}
	orig, err = yaml.YAMLToJSON(raw)
	if err != nil {
		return fmt.Errorf("could not parse input values.yaml YAML into JSON: %w", err)
	}

	// Keep a copy of the original to diff later.
	transformed = orig

	for start, end := range getKeyReMaps() {
		transformed, err = Move(start, end, transformed)
		if err != nil {
			logger.Error(err, "migration failed")
		}
	}

	fmt.Printf("\n%s\n", string(transformed))

	return nil
}

func Move(start, end string, doc []byte) ([]byte, error) {
	found := gjson.GetBytes(doc, start)
	if !found.Exists() {
		return doc, fmt.Errorf("key %s for target %s not found", start, end)
	}

	result, err := sjson.SetBytes(doc, end, found.Value())
	if err != nil {
		return doc, fmt.Errorf("could not inject %s at %s: %w", start, end, err)
	}

	return result, nil
}
