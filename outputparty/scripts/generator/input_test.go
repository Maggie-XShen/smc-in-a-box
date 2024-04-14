package generator_test

import (
	"testing"
	"time"

	"example.com/SMC/outputparty/scripts/generator"
)

func TestGenerateServerInput(t *testing.T) {
	generator.GenerateOPInput(2, time.Now(), 8, "./input")
}
