package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
)

func main() {

	n_s := flag.Int("n", 1, "number of servers")
	flag.Parse()

	// Configure servers
	/**
	var ports []string
	for i := 0; i < *n_s; i++ {
		port := fmt.Sprintf("808%s", strconv.Itoa(i))
		ports = append(ports, port)
	}
	generator.GenerateServerConfig(*n_s, ports, "../config_generator/server_template.json", "generator/examples/")
	**/

	// Start the servers
	var processes []*exec.Cmd
	for i := 1; i <= *n_s; i++ {
		conf_path := fmt.Sprintf("-confpath=../config_generator/examples/config_s%s.json", strconv.Itoa(i))
		server := startServer("../cmd/server", conf_path)
		processes = append(processes, server)
	}

	for _, process := range processes {
		process.Wait()
	}

	// Stop the servers
	for _, process := range processes {
		stop(process)
	}

	log.Println("All server have finished.")

}

func startServer(serverCmd string, arg string) *exec.Cmd {
	cmd := exec.Command(serverCmd, arg)
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
