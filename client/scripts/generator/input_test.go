package generator_test

import (
	"testing"

	"example.com/SMC/client/scripts/generator"
)

func TestGenerateInput(t *testing.T) {
	generator.GenerateClientInput(6, 2, []int{1, 1}, "./input")
}
