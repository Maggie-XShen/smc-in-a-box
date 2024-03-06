package generator_test

import (
	"testing"

	"example.com/SMC/outputparty/scripts/generator"
)

func TestGenerateServerInput(t *testing.T) {
	generator.GenerateOPInput(2, "2024-03-06 19:45:00", 8, "./input")
}
