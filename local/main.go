package main

import (
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

	n_client := 1
	n_client_mal := 1
	n_exp := 1
	n_input := []int{10000} //clarify the number of inputs for each experiment, e.g. n_input={1,2} means first experiment has 1 input, second experiment has 2 inputs

	n_server := 4
	server_port := []string{"50001", "50002", "50003", "50004", "50005", "50006", "50007"} //port for each server

	n_outputparty := 1
	op_port := []string{"60000"}

	start := time.Now().UTC()
	clientShareDue := start.Add(time.Minute * 3)
	t1 := 4  // ComplaintDue = ClientShareDue + t1
	t2 := 8  // MaskedShareDue = ClientShareDue + t2
	t3 := 11 // ServerShareDue = ClientShareDue + t3

	client_gen.GenerateClientConfig(n_client, "client_template.json", "./client_config")

	client_gen.GenerateClientInput(n_client, n_exp, n_input, "./client_input")

	server_gen.GenerateServerConfigLocal(n_server, server_port[:n_server], "server_template.json", "./server_config")

	server_gen.GenerateServerInput(n_exp, clientShareDue, t1, t2, "http://127.0.0.1:60000/serverShare/", "./server_input")

	output_gen.GenerateOPConfig(n_outputparty, op_port, "outputparty_template.json", "./op_config")

	output_gen.GenerateOPInput(n_exp, clientShareDue, t3, "./op_input")

	run(n_server, n_outputparty, n_client, n_client_mal)

}

func run(n_server, n_outputparty, n_client, n_client_mal int) {

	l1 := n_server
	firstGroup := make([][]string, l1)
	for i := 0; i < l1; i++ {
		firstGroup[i] = make([]string, 7)
		firstGroup[i][0] = "../server/cmd/cmd"
		firstGroup[i][1] = fmt.Sprintf("-confpath=./server_config/config_s%s.json", strconv.Itoa(i+1))
		firstGroup[i][2] = "-inputpath=./server_input/experiments.json"
		firstGroup[i][3] = "-mode=http"
		firstGroup[i][4] = "-logpath=./server_log/"
		firstGroup[i][5] = fmt.Sprintf("-n_client=%d", n_client)
		firstGroup[i][6] = fmt.Sprintf("-n_client_mal=%d", n_client_mal)
	}

	secondGroup := make([][]string, n_outputparty)
	for i := 0; i < n_outputparty; i++ {
		secondGroup[i] = make([]string, 6)
		secondGroup[i][0] = "../outputparty/cmd/cmd"
		secondGroup[i][1] = fmt.Sprintf("-confpath=./op_config/config_op%s.json", strconv.Itoa(i+1))
		secondGroup[i][2] = "-inputpath=./op_input/experiments.json"
		secondGroup[i][3] = "-mode=http"
		secondGroup[i][4] = "-logpath=./op_log/"
		secondGroup[i][5] = fmt.Sprintf("-n_client=%d", n_client)
	}

	thirdGroup := make([][]string, n_client)

	for i := 0; i < n_client_mal; i++ {
		thirdGroup[i] = make([]string, 5)
		thirdGroup[i][0] = "../client/cmd/cmd"
		thirdGroup[i][1] = fmt.Sprintf("-confpath=./client_config/config_c%s.json", strconv.Itoa(i+1))
		thirdGroup[i][2] = fmt.Sprintf("-inputpath=./client_input/input_c%s.json", strconv.Itoa(i+1))
		thirdGroup[i][3] = "-logpath=./client_log/"
		thirdGroup[i][4] = "-mode=malicious"
	}

	for i := n_client_mal; i < n_client-n_client_mal; i++ {
		thirdGroup[i] = make([]string, 5)
		thirdGroup[i][0] = "../client/cmd/cmd"
		thirdGroup[i][1] = fmt.Sprintf("-confpath=./client_config/config_c%s.json", strconv.Itoa(i+1))
		thirdGroup[i][2] = fmt.Sprintf("-inputpath=./client_input/input_c%s.json", strconv.Itoa(i+1))
		thirdGroup[i][3] = "-logpath=./client_log/"
		thirdGroup[i][4] = "-mode=honest"
	}

	var wg sync.WaitGroup

	// Execute servers in parallel
	for _, cmd := range firstGroup {
		wg.Add(1)
		go executeFirstGroup(cmd[0], cmd[1], cmd[2], cmd[3], cmd[4], cmd[5], cmd[6], &wg)
	}

	// Execute output parties in parallel
	for _, cmd := range secondGroup {
		wg.Add(1)
		go executeSecondGroup(cmd[0], cmd[1], cmd[2], cmd[3], cmd[4], cmd[5], &wg)
	}

	time.Sleep(30 * time.Second)

	for _, cmd := range thirdGroup {
		wg.Add(1)
		go executeThirdGroup(cmd[0], cmd[1], cmd[2], cmd[3], cmd[4], &wg)
	}

	// Wait for all commands to finish
	wg.Wait()
}

func executeFirstGroup(command, conf_path, input_path, mode, log_path, n_client, n_client_mal string, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command(command, conf_path, input_path, mode, log_path, n_client, n_client_mal)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error executing command %s with %s: %v\n", command, conf_path, err)
		return
	}
}

func executeSecondGroup(command, conf_path, input_path, mode, log_path, n_client string, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command(command, conf_path, input_path, mode, log_path, n_client)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error executing command %s with %s: %v\n", command, conf_path, err)
		return
	}
}

func executeThirdGroup(command, conf_path, input_path, log_path, mode string, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command(command, conf_path, input_path, log_path, mode)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error executing command %s with %s: %v\n", command, conf_path, err)
		return
	}
}
