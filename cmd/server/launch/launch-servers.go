package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"example.com/SMC/cmd/server/config"
)

func main() {

	n_s := flag.Int("n", 1, "number of servers")
	flag.Parse()

	// Configure servers
	var ports []string
	for i := 0; i < *n_s; i++ {
		port := fmt.Sprintf("808%s", strconv.Itoa(i))
		ports = append(ports, port)
	}
	config.GenerateServerConfig(*n_s, ports, "../config/server_template.json", "../config/examples/")

	// Start the servers
	var processes []*exec.Cmd
	for i := 1; i <= *n_s; i++ {
		conf_path := fmt.Sprintf("-confpath=../config/examples/config_s%s.json", strconv.Itoa(i))
		data_path := "-datapath=../data.json"
		server := startServer("../server", conf_path, data_path)
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
