package generator_test

import (
	"testing"

	"example.com/SMC/client/scripts/generator"
)

func TestGenerateGonfig(t *testing.T) {
	generator.GenerateClientConfig(6, "client_template.json", "./config")
}
