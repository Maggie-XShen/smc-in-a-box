package config_generator_test

import (
	"testing"

	"example.com/SMC/outputparty/scripts/config_generator"
)

func TestGenerateGonfig(t *testing.T) {
	config_generator.GenerateOPConfig(2, "outputparty_template.json", "examples/")
}
