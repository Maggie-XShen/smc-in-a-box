package generator_test

import (
	"testing"

	"example.com/SMC/server/scripts/generator"
)

func TestGenerateServerInput(t *testing.T) {
	generator.GenerateServerInput(2, "2024-03-06 19:45:00", 2, 5, "http://127.0.0.1:50000/serverShare/", "./input")
}
