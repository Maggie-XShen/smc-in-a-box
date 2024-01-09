# Simple Prototype System for Private Statistics
A simple prototype system for private statistics,  comprising clients, several secure multiparty computation servers, and output party(e.g., data analyst) written in Golang.

The prototype system simulates a scenario in which the output party designs experiments or conducts surveys to collect statistics, such as a data analyst. In this scenario, the client actively participates in the experiment or survey. Servers play a crucial role in assisting the output party in obtaining statistics without learning the client's input, ensuring the privacy of the client's data remains protected.

## Client
To run a single client instance, first go to the folder client/cmd and run command "go build" to compile the program, then at the same folder run command
```
./cmd -confpath=“path_to_client_config_file” -inputpath=“path_to_client_input_file”
```

If confpath option is not provided, the program will read the config file at client/config/client.json

If inputpath option is not provided, the program will read the client input from client/cmd/input.json

To launch several client instances at once and use same client input, first make sure client executable is in the folder client/cmd, then go to the folder client/scripts/launch and run command
```
go run launch-clients.go -n=3
```
where n is the number of client instances
  
Clients config files will be generated automatically and stored in the folder client/scripts/condig_generator/examples

## Server
In dev2.0, it is assumed that each server has pre-existing experiments data and registered clients information. These data are stored at server/cmd/registry.json

To run a single server instance, go to the folder server/cmd and and run command "go build" to compile the program, then at the same folder run command
```
./cmd -confpath="path_to_server_config_file" -registrypath="path_to_registry_file"
```
If confpath option is not provided, the program will read the config file at server/config/server.json

If registrypath option is not provided, the program will read the experiments information and registered clients information from server/cmd/registry.json

## Output Party
To run a single output party instance, go to the folder outputparty/cmd and run command "go build" to compile the program, then at the same folder run command
```
./cmd -confpath="path_to_output_party_config_file" -exppath="path_to_experiments_file"
```
If confpath option is not provided, the program will read output party configuration from outputparty/config/outputparty.json

If registrypath option is not provided, the program will read the experiments information from outputparty/cmd/experiments.json

## Note
The third version of development moves to use replicated secret-sharing as building block compared to the second version of development using shamir secret-sharing. ZK proof protocol is changed to non-interactive version using fiat shamir transformation.
  
## To Do
+ Database needs to change per confirguration