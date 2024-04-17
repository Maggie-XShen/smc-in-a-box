package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	client_gen "example.com/SMC/client/scripts/generator"
	output_gen "example.com/SMC/outputparty/scripts/generator"
	server_gen "example.com/SMC/server/scripts/generator"
)

func main() {
	party := flag.String("party", "", "type of instance")
	start := flag.String("start", "", "start time of server")
	flag.Parse()

	n_client := 10
	n_exp := 1
	n_input := []int{10} //clarify the number of inputs for each experiment, e.g. n_input={1,2} means first experiment has 1 input, second experiment has 2 inputs

	n_server := 4
	server_port := []string{"https://smc-server-1.cs-georgetown.net:", "https://smc-server-2.cs-georgetown.net:", "https://smc-server-3.cs-georgetown.net:", "https://smc-server-4.cs-georgetown.net:", "https://smc-server-5.cs-georgetown.net:", "https://smc-server-6.cs-georgetown.net:", "https://smc-server-7.cs-georgetown.net:", "https://smc-server-8.cs-georgetown.net:", "https://smc-server-9.cs-georgetown.net:", "https://smc-server-10.cs-georgetown.net:"} //url for each server

	n_outputparty := 1
	op_port := []string{"443"}

	timestamp, _ := time.Parse("2006-01-02 15:04:05.999999999 +0000 UTC", *start)
	clientShareDue := timestamp.Add(time.Minute * 4)
	t1 := 2  // ComplaintDue = ClientShareDue + t1
	t2 := 5  // MaskedShareDue = ClientShareDue + t2
	t3 := 10 // ServerShareDue = ClientShareDue + t3

	if *party == "client" {

		client_gen.GenerateClientConfig(n_client, "client_template.json", "./client_config")

		client_gen.GenerateClientInput(n_client, n_exp, n_input, "./client_input")

		run(n_client)

	} else if *party == "server" {
		server_gen.GenerateServerConfigCloud(n_server, server_port[:n_server], "server_template.json", "./server_config")

		server_gen.GenerateServerInput(n_exp, clientShareDue, t1, t2, "https://smc-outputparty.cs-georgetown.net:443/serverShare/", "./server_input")

		arg := make([]string, 5)
		arg[0] = "../server/cmd/cmd"
		arg[1] = "-confpath=./server_config/config_server%s.json"
		arg[2] = "-inputpath=./server_input/experiments.json"
		arg[3] = "-logpath=./server_log/"
		arg[4] = fmt.Sprintf("-n_client=%d", n_client)

		cmd := exec.Command(arg[0], arg[1], arg[2], arg[3], arg[4])

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error executing command %s with %s: %v\n", arg[0], arg[1], err)
			return
		}

	} else if *party == "outputparty" {
		output_gen.GenerateOPConfig(n_outputparty, op_port, "outputparty_template.json", "./op_config")

		output_gen.GenerateOPInput(n_exp, clientShareDue, t3, "./op_input")

		arg := make([]string, 5)
		arg[0] = "../outputparty/cmd/cmd"
		arg[1] = "-confpath=./op_config/config_op.json"
		arg[2] = "-inputpath=./op_input/experiments.json"
		arg[3] = "-logpath=./op_log/"
		arg[4] = fmt.Sprintf("-n_client=%d", n_client)

		cmd := exec.Command(arg[0], arg[1], arg[2], arg[3], arg[4])

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error executing command %s with %s: %v\n", arg[0], arg[1], err)
			return
		}
	}

}

func run(n_client int) {
	thirdGroup := make([][]string, n_client)
	for i := 0; i < n_client; i++ {
		thirdGroup[i] = make([]string, 4)
		thirdGroup[i][0] = "../client/cmd/cmd"
		thirdGroup[i][1] = fmt.Sprintf("-confpath=./client_config/config_c%s.json", strconv.Itoa(i+1))
		thirdGroup[i][2] = fmt.Sprintf("-inputpath=./client_input/input_c%s.json", strconv.Itoa(i+1))
		thirdGroup[i][3] = "-logpath=./client_log/"

	}

	var wg sync.WaitGroup

	for _, cmd := range thirdGroup {
		wg.Add(1)
		go executeSecondGroup(cmd[0], cmd[1], cmd[2], cmd[3], &wg)
	}

	// Wait for all commands to finish
	wg.Wait()
}

func executeSecondGroup(command, conf_path, input_path, log_path string, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command(command, conf_path, input_path, log_path)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error executing command %s with %s: %v\n", command, conf_path, err)
		return
	}
}
