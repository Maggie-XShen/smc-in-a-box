package generator_test

import (
	"testing"

	"example.com/SMC/outputparty/scripts/generator"
)

func TestGenerateGonfig(t *testing.T) {
	generator.GenerateOPConfig(1, "outputparty_template.json", "./config")
}
