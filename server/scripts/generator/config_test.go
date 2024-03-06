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
	ips := []string{"https://smc-server1.cs-georgetown.net:", "https://smc-server2.cs-georgetown.net:", "https://smc-server3.cs-georgetown.net:", "https://smc-server4.cs-georgetown.net:", "https://smc-server5.cs-georgetown.net:", "https://smc-server6.cs-georgetown.net:"}
	generator.GenerateServerConfigCloud(6, ips, "server_template.json", "./config/")
}
