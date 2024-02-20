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

	n_c := flag.Int("n", 1, "number of clients")
	flag.Parse()

	fmt.Printf("%d\n", *n_c)

	// Configure clients
	//config_generator.GenerateClientConfig(*n_c, "../config_generator/client_template.json", "../../config/examples/")

	// Start the clients
	var processes []*exec.Cmd
	input_path := "-inputpath=../../cmd/input.json"
	for i := 1; i <= *n_c; i++ {
		conf_path := fmt.Sprintf("-confpath=../../config/examples/config_c%s.json", strconv.Itoa(i))

		client := startClient("../../cmd/cmd", conf_path, input_path)
		processes = append(processes, client)
	}

	for _, process := range processes {
		process.Wait()
	}

	log.Println("All clients have finished.")

}

func startClient(clientCmd string, arg1 string, arg2 string) *exec.Cmd {
	cmd := exec.Command(clientCmd, arg1, arg2)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	return cmd
}
