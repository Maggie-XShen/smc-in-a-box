package config_test

import (
	"testing"

	"example.com/SMC/cmd/client/config"
)

func TestGenerateGonfig(t *testing.T) {
	config.GenerateClientConfig(2, "client_template.json", "examples/")
}
