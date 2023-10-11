package config_generator_test

import (
	"testing"

	"example.com/SMC/server/scripts/config_generator"
)

func TestGenerateGonfig(t *testing.T) {
	ports := []string{"8081", "8082", "8083"}
	config_generator.GenerateServerConfig(3, ports, "server_template.json", "examples/")
}
