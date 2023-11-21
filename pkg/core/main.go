package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"sigs.k8s.io/yaml"
)

// TODO a proper type would be nice, but I don't think we can create a perfect
// one. aside from annoyances where some keys are commented by default,
// values.yaml is not a regular struct. Some sections like .env allow for
// arbitrary keys. We can presumably map[string]interface{} these, however.
// Unsure if there's a benefit to a struct beyond just map[string]interface{}
// everything and assuming found keys contain valid values. This is reasonable,
// since invalid values wouldn't have rendered in Helm.
//
// the gjson/sjson approach below may well be better for a limited set of keys,
// since most everything we want to just leave in place, and it does that. structs
// would probably be more difficult to maintain over time--the pure JSON approach
// doesn't require a rigid schema, just a basic "expect key here" paths. we can
// bundle sets of paths into migrations releases similar to Kong's to provide a
// path between versions, e.g. you could run a "foo.red -> foo.blue" and then
// "foo.blue -> bar.blue" after. As long as they run in order, they should reach
// the final state without returning full structs for each version.

// TODO some limitations of this approach:
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

// TODO this is a basic single set, random order collection of migrations, since there's a readily available type for
// that. should probably convert to a struct we can stick in an ordered array and then allow running different sets
// in the order specified, to handle migrations that depend on one another.

// TODO there are some keys that only migrate if you're coming from the ingress chart because they were using existing
// global settings to modify the controller deployment. AFAIK podAnnotations and proxy are the only two of these. proxy
// shouldn't matter since we can infer the correct setting in the split kong chart automatically.
// controller.podAnnotations would move to ingressController.deployment.pod.annotations, but building a second set of
// conditional maps complicates things. We may just ignore this and use the default annotations in kong 3.x to handle
// most users and ask the remainder to migrate manually

// TODO the kong chart is sort of affected by a concern similar to the above: annotations previously applied to the
// single Deployment via podAnnotations and the like _may_ be relevant for the new controller Deployment, but there's
// no way to know whether they were in place for the controller or for the proxy

// getKeyReMaps returns a map of strings to strings. Keys are the original locations of a key in values.yaml and values
// are their new locations. Both are in dotted string format: "foo.bar.baz" indicates a YAML structure like:
// foo:
//
//	bar:
//	  baz: {}

func getControllerKeys() map[string]string {
	return map[string]string{
		"ingressController.image":          "ingressController.deployment.pod.container.image",
		"ingressController.args":           "ingressController.deployment.pod.container.args",
		"ingressController.env":            "ingressController.deployment.pod.container.env",
		"ingressController.customEnv":      "ingressController.deployment.pod.container.customEnv",
		"ingressController.livenessProbe":  "ingressController.deployment.pod.container.livenessProbe",
		"ingressController.readinessProbe": "ingressController.deployment.pod.container.readinessProbe",
		"ingressController.resources":      "ingressController.deployment.pod.container.resources",
	}
}

func getGatewayKeys() map[string]string {
	return map[string]string{}
}

// TODO weird keys
// proxy.* probably no longer necessary since it's only required to deal with the subchart boundary

func getIngressControllerKeys() map[string]string {
	return map[string]string{
		"controller.ingressController.image":          "ingressController.deployment.pod.container.image",
		"controller.ingressController.args":           "ingressController.deployment.pod.container.args",
		"controller.ingressController.env":            "ingressController.deployment.pod.container.env",
		"controller.ingressController.customEnv":      "ingressController.deployment.pod.container.customEnv",
		"controller.ingressController.livenessProbe":  "ingressController.deployment.pod.container.livenessProbe",
		"controller.ingressController.readinessProbe": "ingressController.deployment.pod.container.readinessProbe",
		"controller.ingressController.resources":      "ingressController.deployment.pod.container.resources",
	}
}

func getIngressGatewayKeys() map[string]string {
	return map[string]string{}
}

type mapFunc func() map[string]string

const (
	controllerPrefix = "controller"
	gatewayPrefix    = "gateway"
	kongChart        = "kong"
	ingressChart     = "ingress"
)

type IngressValues struct {
	Gateway    map[string]interface{} `json:"gateway,omitempty" yaml:"gateway,omitempty"`
	Controller map[string]interface{} `json:"controller,omitempty" yaml:"controller,omitempty"`
}

func Merge(_ context.Context, c *Config, logger logr.Logger) error {
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
	_, err = input.Read(raw)
	if err != nil {
		return fmt.Errorf("could not read input values.yaml: %w", err)
	}
	// for whatever reason attempting to directly unmarshal from YAML results in an empty object
	jsoned, err := yaml.YAMLToJSON(raw)
	if err != nil {
		return fmt.Errorf("could not parse input values.yaml YAML into JSON: %w", err)
	}
	var orig IngressValues
	transformed := make(map[string]interface{})
	err = json.Unmarshal(jsoned, &orig)
	if err != nil {
		return fmt.Errorf("could not parse input values.yaml: %w", err)
	}

	delete(orig.Gateway, "ingressController")
	delete(orig.Gateway, "enabled")
	delete(orig.Controller, "deployment")
	delete(orig.Controller, "enabled")
	delete(orig.Controller, "proxy")

	for key, value := range orig.Gateway {
		transformed[key] = value
	}

	for key, value := range orig.Controller {
		if _, exists := transformed[key]; exists {
			fmt.Println(fmt.Sprintf("key %s exists in both gateway and controller, using gateway", key))
		} else {
			transformed[key] = value
		}
	}

	yamlOut, err := yaml.Marshal(transformed)
	if err != nil {
		return fmt.Errorf("could not marshal YAML: %w", err)
	}
	fmt.Printf("\n%s\n", string(yamlOut))
	return nil
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

	remaps := map[string]mapFunc{
		controllerPrefix: getControllerKeys,
		gatewayPrefix:    getGatewayKeys,
	}

	for _, prefix := range []string{controllerPrefix, gatewayPrefix} {
		for start, end := range remaps[prefix]() {
			fullStart := start
			if c.SourceChart == ingressChart {
				fullStart = fmt.Sprintf("%s.%s", prefix, start)
			}
			transformed, err = Move(fullStart, end, transformed)
			if err != nil {
				logger.Error(err, "migration failed")
			}
			// not immediately clear why, but attempting to move AND delete within Move (the contents of Delete originally
			// followed the sjson.SetBytes() call and error check) resulted in it deleting both the old and new key.
			// Presumably something about how it addresses the values internally. Returning and then deleting avoids this,
			// since we have a new []byte to work with.
			transformed, err = Delete(fullStart, transformed)
			if err != nil {
				logger.Error(err, "cleanup failed")
			}
		}
	}

	if c.OutputFormat == "json" {
		fmt.Printf("\n%s\n", string(transformed))
	} else {
		yamlOut, err := yaml.JSONToYAML(transformed)
		if err != nil {
			return fmt.Errorf("could not convert back to YAML: %w", err)
		}
		fmt.Printf("\n%s\n", string(yamlOut))
	}

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

func Delete(key string, doc []byte) ([]byte, error) {
	result, err := sjson.DeleteBytes(doc, key)
	if err != nil {
		return doc, fmt.Errorf("could not delete old content at %s: %w", key, err)
	}
	return result, nil
}
