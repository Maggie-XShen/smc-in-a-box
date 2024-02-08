package config_generator_test

import (
	"testing"

	"example.com/SMC/server/scripts/config_generator"
)

func TestGenerateGonfig(t *testing.T) {
	ports := []string{"50001", "50002", "50003", "50004", "50005", "50006"}
	config_generator.GenerateServerConfig(6, ports, "server_template.json", "../../config/examples/")
}
