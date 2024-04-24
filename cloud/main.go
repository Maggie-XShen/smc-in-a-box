package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	client_gen "example.com/SMC/client/scripts/generator"
	output_gen "example.com/SMC/outputparty/scripts/generator"
	server_gen "example.com/SMC/server/scripts/generator"
)

func main() {
	party := flag.String("party", "", "type of instance")
	sid := flag.Int("sid", 0, "server id") //e.g. 0 for s0
	client_threads := flag.Int("client_threads", 0, "number of clients running on the same machine")
	start_cid := flag.Int("start_cid", 0, "client id") //e.g. 0 for c0
	n_server := flag.Int("n_server", 0, "number of servers")
	n_clients := flag.Int("n_clients", 0, "number of total clients")
	n_clients_mal := flag.Int("n_clients_mal", 0, "number of malicious clients")
	n_input := flag.Int("n_input", 0, "number of inputs")
	template_path := flag.String("template_path", "", "path to config template")
	start := flag.String("start", "", "start time of server")
	d0 := flag.Int("d0", 0, "duration from start time to client share due")
	d1 := flag.Int("d1", 0, "duration from client share due to complaint due")
	d2 := flag.Int("d2", 0, "duration from client share due to masked share due")
	d3 := flag.Int("d3", 0, "duration from client share due to server share due")

	flag.Parse()

	n_exp := 1
	input_list := []int{*n_input} //clarify the number of inputs for each experiment, e.g. n_input={1,2} means first experiment has 1 input, second experiment has 2 inputs

	server_port := []string{"https://server0.privatestats.org:", "https://server1.privatestats.org:", "https://server2.privatestats.org:", "https://server3.privatestats.org:", "https://server4.privatestats.org:", "https://server5.privatestats.org:", "https://server6.privatestats.org:", "https://server7.privatestats.org:", "https://server8.privatestats.org:", "https://server9.privatestats.org:"} //url for each server

	n_outputparty := 1
	op_port := []string{"443"}

	timestamp, _ := time.Parse("2006-01-02 15:04:05.999999999 +0000 UTC", *start)

	t0 := time.Duration(*d0)
	clientShareDue := timestamp.Add(time.Minute * t0)
	t1 := *d1 // ComplaintDue = ClientShareDue + t1
	t2 := *d2 // MaskedShareDue = ClientShareDue + t2
	t3 := *d3 // ServerShareDue = ClientShareDue + t3

	if *party == "client" {

		client_gen.GenerateClientConfigCloud(*n_clients, filepath.Join(*template_path, "client_template.json"), "./client_config")

		client_gen.GenerateClientInputCloud(*n_clients, n_exp, input_list, "./client_input")

		run(*client_threads, *start_cid)

	} else if *party == "client_mal" {

		client_gen.GenerateClientConfigCloud(*n_clients, filepath.Join(*template_path, "client_template.json"), "./client_config")

		client_gen.GenerateClientInputCloud(*n_clients, n_exp, input_list, "./client_input")

		run_mal(*n_clients_mal, *start_cid)

	} else if *party == "server" {
		server_gen.GenerateServerConfigCloud(*n_server, server_port[:*n_server], filepath.Join(*template_path, "server_template.json"), "./server_config")

		server_gen.GenerateServerInput(n_exp, clientShareDue, t1, t2, "https://outputparty.privatestats.org/serverShare/", "./server_input")

		arg := make([]string, 6)
		arg[0] = "../server/cmd/cmd"
		arg[1] = fmt.Sprintf("-confpath=./server_config/config_s%s.json", strconv.Itoa(*sid))
		arg[2] = "-inputpath=./server_input/experiments.json"
		arg[3] = "-logpath=./server_log/"
		arg[4] = fmt.Sprintf("-n_client=%d", n_clients)
		arg[5] = fmt.Sprintf("-n_client_mal=%d", n_clients_mal)

		cmd := exec.Command(arg[0], arg[1], arg[2], arg[3], arg[4], arg[5])

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error executing command %s with %s: %v\n", arg[0], arg[1], err)
			return
		}

	} else if *party == "outputparty" {
		output_gen.GenerateOPConfig(n_outputparty, op_port, filepath.Join(*template_path, "outputparty_template.json"), "./op_config")

		output_gen.GenerateOPInput(n_exp, clientShareDue, t3, "./op_input")

		arg := make([]string, 5)
		arg[0] = "../outputparty/cmd/cmd"
		arg[1] = "-confpath=./op_config/config_op1.json"
		arg[2] = "-inputpath=./op_input/experiments.json"
		arg[3] = "-logpath=./op_log/"
		arg[4] = fmt.Sprintf("-n_client=%d", n_clients)

		cmd := exec.Command(arg[0], arg[1], arg[2], arg[3], arg[4])

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error executing command %s with %s: %v\n", arg[0], arg[1], err)
			return
		}
	}

}

func run(client_threads int, start_cid int) {
	thirdGroup := make([][]string, client_threads)
	cid := start_cid
	for i := 0; i < client_threads; i++ {
		thirdGroup[i] = make([]string, 5)
		thirdGroup[i][0] = "../client/cmd/cmd"
		thirdGroup[i][1] = fmt.Sprintf("-confpath=./client_config/config_c%s.json", strconv.Itoa(cid))
		thirdGroup[i][2] = fmt.Sprintf("-inputpath=./client_input/input_c%s.json", strconv.Itoa(cid))
		thirdGroup[i][3] = "-logpath=./client_log/"
		thirdGroup[i][4] = "-mode=honest"
		cid++
	}

	var wg sync.WaitGroup

	for _, cmd := range thirdGroup {
		wg.Add(1)
		go executeSecondGroup(cmd[0], cmd[1], cmd[2], cmd[3], cmd[4], &wg)
	}

	// Wait for all commands to finish
	wg.Wait()
}

func run_mal(n_client_mal int, start_cid int) {
	thirdGroup := make([][]string, n_client_mal)
	cid := start_cid
	for i := 0; i < n_client_mal; i++ {
		thirdGroup[i] = make([]string, 4)
		thirdGroup[i][0] = "../client/cmd/cmd"
		thirdGroup[i][1] = fmt.Sprintf("-confpath=./client_config/config_c%s.json", strconv.Itoa(cid))
		thirdGroup[i][2] = fmt.Sprintf("-inputpath=./client_input/input_c%s.json", strconv.Itoa(cid))
		thirdGroup[i][3] = "-logpath=./client_log/"
		cid++
	}

	var wg sync.WaitGroup

	for _, cmd := range thirdGroup {
		wg.Add(1)
		go executeSecondGroup(cmd[0], cmd[1], cmd[2], cmd[3], cmd[4], &wg)
	}

	// Wait for all commands to finish
	wg.Wait()
}

func executeSecondGroup(command, conf_path, input_path, log_path, mode string, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command(command, conf_path, input_path, log_path, mode)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error executing command %s with %s: %v\n", command, conf_path, err)
		return
	}
}
