package config_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/config"
)

var configExamplePath string

func init() {
	_, currentFile, _, _ := runtime.Caller(0)
	configExamplePath = filepath.Join(filepath.Dir(currentFile), "..", "..", "configs", "config.toml.example")
}

func TestParse(t *testing.T) {
	cfg, err := config.Parse(configExamplePath)
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Log.Level)
}
