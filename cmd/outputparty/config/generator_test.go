package config_test

import (
	"testing"

	"example.com/SMC/cmd/outputparty/config"
)

func TestGenerateGonfig(t *testing.T) {
	config.GenerateOPConfig(1, "outputparty_template.json", "examples/")
}
