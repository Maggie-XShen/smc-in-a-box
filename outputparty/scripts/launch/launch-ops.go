package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"example.com/SMC/outputparty/scripts/generator"
)

func main() {

	n_op := flag.Int("n", 1, "number of output parties")
	flag.Parse()

	// Configure output party
	generator.GenerateOPConfig(*n_op, []string{"60000"}, "../generator/outputparty_template.json", "../generator/config")

	// Start the output party
	var processes []*exec.Cmd
	for i := 1; i <= *n_op; i++ {
		conf_path := fmt.Sprintf("-confpath=../generator/config/config_op%s.json", strconv.Itoa(i))
		exp_path := "-exppath=../generator/input/experiments.json"
		outputparty := startOP("../../cmd/cmd", conf_path, exp_path)
		processes = append(processes, outputparty)
	}

	for _, process := range processes {
		process.Wait()
	}

	for _, process := range processes {
		stop(process)
	}
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
