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
	confpath := flag.String("confpath", "", "config file path")
	inputpath := flag.String("inputpath", "", "experiments file path")
	flag.Parse()

	// Start the clients
	var processes []*exec.Cmd

	for i := 1; i <= *n_c; i++ {
		arg1 := fmt.Sprintf("-confpath="+*confpath+"config_c%s.json", strconv.Itoa(i))
		arg2 := fmt.Sprintf("-inputpath"+*inputpath+"input_c%s.json", strconv.Itoa(i))
		client := startClient("../../cmd/cmd", arg1, arg2)
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
