package config_test

import (
	"testing"

	"example.com/SMC/cmd/server/config"
)

func TestGenerateGonfig(t *testing.T) {
	urls := []string{"8080", "8081", "8082"}
	config.GenerateServerConfig(3, urls, "server_template.json", "examples/")
}
