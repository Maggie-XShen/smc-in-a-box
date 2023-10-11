## Client
+ To run a single client instance, go to the folder client/cmd,
run ./cmd -confpath=“path_to_client_config_file” -inputpath=“path_to_client_input_file”
If confpath option is not provided, the program will read the config file at client/config/client.json

If inputpath option is not provided, the program will read the input file at client/cmd/input.json

+ To launch several client instances at once, go to the folder client/scripts/launch, run  go run launch-clients.go -n=3,  where n is the number of client instances
  
Clients config files will be generated automatically and stored in the folder client/scripts/condig_generator/examples

## Server

## Output Party

  
