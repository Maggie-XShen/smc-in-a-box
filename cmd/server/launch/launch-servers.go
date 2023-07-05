package main

import (
	"log"
	"os"
	"os/exec"

	"example.com/SMC/cmd/server/config"
)

func main() {
	/**
	ns := flag.Int("n", 1, "number of servers")
	flag.Parse()**/

	// Configure servers
	urls := []string{"8080", "8081", "8082"}
	config.GenerateServerConfig(3, urls, "../config/server_template.json", "../config/examples/")

	// Start the servers
	server1 := startServer("../server", "-confpath=../config/examples/config_s1.json", "-datapath=../tables.json")
	server2 := startServer("../server", "-confpath=../config/examples/config_s2.json", "-datapath=../tables.json")
	server3 := startServer("../server", "-confpath=../config/examples/config_s3.json", "-datapath=../tables.json")

	server1.Wait()
	server2.Wait()
	server3.Wait()

	// Stop the servers
	stop(server1)
	stop(server2)
	stop(server3)

	log.Println("All server have finished.")

}

func startServer(serverCmd string, arg1 string, arg2 string) *exec.Cmd {
	cmd := exec.Command(serverCmd, arg1, arg2)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	return cmd
}

func stop(cmd *exec.Cmd) {
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		log.Fatalf("Failed to stop process: %v", err)
	}
}
