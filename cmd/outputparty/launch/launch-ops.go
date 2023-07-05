package main

import (
	"log"
	"os"
	"os/exec"

	"example.com/SMC/cmd/outputparty/config"
)

func main() {
	/**
	nop := flag.Int("n", 1, "number of output parties")
	flag.Parse()**/

	// Configure output party
	config.GenerateOPConfig(1, "../config/outputparty_template.json", "../config/examples/")

	// Start the output party
	op1 := startOP("../outputparty", "-confpath=../config/examples/config_op1.json", "-exppath=../experiments_data.json")
	op1.Wait()

	stop(op1)
}

func startOP(opCmd string, arg1 string, arg2 string) *exec.Cmd {
	cmd := exec.Command(opCmd, arg1, arg2)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start output party: %v", err)
	}
	return cmd
}

func stop(cmd *exec.Cmd) {
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		log.Fatalf("Failed to stop process: %v", err)
	}
}
