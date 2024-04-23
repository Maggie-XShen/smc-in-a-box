package generator_test

import (
	"testing"

	"example.com/SMC/server/scripts/generator"
)

func TestGenerateGonfigLocal(t *testing.T) {
	ports := []string{"50001", "50002", "50003", "50004", "50005", "50006"}
	generator.GenerateServerConfigLocal(6, ports, "server_template.json", "./config/")
}

func TestGenerateGonfigCloud(t *testing.T) {
	ips := []string{"https://server0.privatestats.org:", "https://server1.privatestats.org:", "https://server2.privatestats.org:", "https://server3.privatestats.org:"}
	generator.GenerateServerConfigCloud(4, ips, "server_template.json", "./config/")
}
