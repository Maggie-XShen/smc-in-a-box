package generator_test

import (
	"testing"
	"time"

	"example.com/SMC/server/scripts/generator"
)

func TestGenerateServerInput(t *testing.T) {
	generator.GenerateServerInput(2, time.Now(), 2, 5, "http://127.0.0.1:50000/serverShare/", "./input")
}
