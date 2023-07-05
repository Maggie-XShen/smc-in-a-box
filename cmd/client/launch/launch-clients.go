package main

import (
	"log"
	"os"
	"os/exec"

	"example.com/SMC/cmd/client/config"
)

func main() {
	/**
	nc := flag.Int("nc", 1, "number of clients")
	flag.Parse()**/

	// Configure clients
	config.GenerateClientConfig(6, "../config/client_template.json", "../config/examples/")

	// Start the clients
	client1 := startClient("../client", "-confpath=../config/examples/config_c1.json")
	client2 := startClient("../client", "-confpath=../config/examples/config_c2.json")
	client3 := startClient("../client", "-confpath=../config/examples/config_c3.json")
	client4 := startClient("../client", "-confpath=../config/examples/config_c4.json")
	client5 := startClient("../client", "-confpath=../config/examples/config_c5.json")
	client6 := startClient("../client", "-confpath=../config/examples/config_c6.json")

	client1.Wait()
	client2.Wait()
	client3.Wait()
	client4.Wait()
	client5.Wait()
	client6.Wait()

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
