package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"sigs.k8s.io/yaml"
)

type IngressValues struct {
	Gateway    map[string]interface{} `json:"gateway,omitempty" yaml:"gateway,omitempty"`
	Controller map[string]interface{} `json:"controller,omitempty" yaml:"controller,omitempty"`
	// IngressController should not exist in normal ingress values, only those that have first passed through the
	// migrate command.
	IngressController map[string]interface{} `json:"ingressController,omitempty" yaml:"ingressController,omitempty"`
}

func MergeOut(ctx context.Context, c *Config, logger logr.Logger) error {
	output, err := Merge(ctx, c, logger)
	if err != nil {
		return err
	}
	fmt.Printf("%s", string(output))
	return nil
}

// Merge combines the "controller" and "gateway" sections of an "ingress" chart values.yaml into a single root-level
// collection of settings for use with the 3.x "kong" chart. It must be run after an initial pass of the main command
// to migrate moved keys to their new locations.
func Merge(_ context.Context, c *Config, logger logr.Logger) (string, error) {
	input, err := os.Open(c.InputFile)
	if err != nil {
		return "", fmt.Errorf("could not open input values.yaml: %w", err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return "", fmt.Errorf("could not inspect input values.yaml: %w", err)
	}

	raw := make([]byte, info.Size())
	_, err = input.Read(raw)
	if err != nil {
		return "", fmt.Errorf("could not read input values.yaml: %w", err)
	}
	// for whatever reason attempting to directly unmarshal from YAML results in an empty object
	jsoned, err := yaml.YAMLToJSON(raw)
	if err != nil {
		return "", fmt.Errorf("could not parse input values.yaml YAML into JSON: %w", err)
	}
	var orig IngressValues
	transformed := make(map[string]interface{})
	err = json.Unmarshal(jsoned, &orig)
	if err != nil {
		return "", fmt.Errorf("could not parse input values.yaml: %w", err)
	}

	delete(orig.Gateway, "ingressController")
	delete(orig.Gateway, "enabled")
	delete(orig.Controller, "enabled")
	delete(orig.Controller, "proxy")

	for key, value := range orig.Gateway {
		transformed[key] = value
	}

	for key, value := range orig.Controller {
		if _, exists := transformed[key]; exists {
			logger.Info("key exists in both gateway and controller, using gateway", "key", key)
		} else {
			transformed[key] = value
		}
	}

	transformed["ingressController"] = orig.IngressController

	yamlOut, err := yaml.Marshal(transformed)
	if err != nil {
		return "", fmt.Errorf("could not marshal YAML: %w", err)
	}
	return string(yamlOut), nil
}
