//go:build integration_tests

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/kong/chart-migrate/pkg/core"
)

func Test_Migrate(t *testing.T) {
	tests := []struct {
		name          string
		sourceFile    string
		expectedFile  string
		expectedState core.Config
	}{
		{
			name:         "001",
			sourceFile:   "testdata/source/001_kong_values.yaml",
			expectedFile: "testdata/expected/001_kong_values.yaml",
		},
		{
			name:         "002",
			sourceFile:   "testdata/source/002_kong_values.yaml",
			expectedFile: "testdata/expected/002_kong_values.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := core.Config{
				SourceChart:  "kong",
				InputFile:    tc.sourceFile,
				OutputFormat: "yaml",
			}

			logger := zapr.NewLogger(zap.NewNop())

			expected, err := os.ReadFile(tc.expectedFile)
			require.NoError(t, err)

			result, err := core.Run(context.Background(), &config, logger)
			require.NoError(t, err)
			require.Equal(t, string(expected), result)
		})
	}
}

func Test_MergeMigrate(t *testing.T) {
	tests := []struct {
		name          string
		sourceFile    string
		expectedFile  string
		expectedState core.Config
	}{
		{
			name:         "001",
			sourceFile:   "testdata/source/001_ingress_values.yaml",
			expectedFile: "testdata/expected/001_ingress_values.yaml",
		},
		{
			name:         "002",
			sourceFile:   "testdata/source/002_ingress_values.yaml",
			expectedFile: "testdata/expected/002_ingress_values.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := core.Config{
				SourceChart:  "ingress",
				InputFile:    tc.sourceFile,
				OutputFormat: "yaml",
			}

			logger := zapr.NewLogger(zap.NewNop())

			expected, err := os.ReadFile(tc.expectedFile)
			require.NoError(t, err)

			migrateResult, err := core.Run(context.Background(), &config, logger)
			require.NoError(t, err)
			migrated, err := os.CreateTemp(os.TempDir(), "chartmigrate.*")
			require.NoError(t, err)
			defer os.Remove(migrated.Name())
			_, err = migrated.Write([]byte(migrateResult))
			require.NoError(t, err)
			require.NoError(t, migrated.Close())

			mergeConfig := core.Config{
				SourceChart:  "ingress",
				InputFile:    migrated.Name(),
				OutputFormat: "yaml",
			}
			result, err := core.Merge(context.Background(), &mergeConfig, logger)
			require.NoError(t, err)
			require.Equal(t, string(expected), result)
		})
	}
}
