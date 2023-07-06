package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"example.com/SMC/cmd/client/config"
)

func main() {

	n_c := flag.Int("n", 1, "number of clients")
	flag.Parse()

	// Configure clients
	config.GenerateClientConfig(*n_c, "../config/client_template.json", "../config/examples/")

	// Start the clients
	var processes []*exec.Cmd
	for i := 1; i <= *n_c; i++ {
		conf_path := fmt.Sprintf("-confpath=../config/examples/config_c%s.json", strconv.Itoa(i))
		client := startClient("../client", conf_path)
		processes = append(processes, client)
	}

	for _, process := range processes {
		process.Wait()
	}

	log.Println("All clients have finished.")

}

func startClient(clientCmd string, arg string) *exec.Cmd {
	cmd := exec.Command(clientCmd, arg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	return cmd
}
