package config_generator_test

import (
	"testing"

	"example.com/SMC/server/scripts/config_generator"
)

func TestGenerateGonfig(t *testing.T) {
	ports := []string{"8081", "8082", "8083", "8084", "8085", "8086"}
	config_generator.GenerateServerConfig(6, ports, "server_template.json", "examples/")
}
