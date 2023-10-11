package config_generator_test

import (
	"testing"

	"example.com/SMC/client/scripts/config_generator"
)

func TestGenerateGonfig(t *testing.T) {
	config_generator.GenerateClientConfig(2, "client_template.json", "examples/")
}
