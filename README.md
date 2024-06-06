# Privacy-Preserving Statistics Collection with Input Validation and Full Security
A practical system for private statistics,  comprising clients, several secure multiparty computation servers, and output party(e.g., data analyst) written in Golang.

The system simulates a scenario in which the output party designs experiments or conducts surveys to collect statistics, such as a data analyst. In this scenario, the client actively participates in the experiment or survey. Servers play a crucial role in assisting the output party in obtaining statistics without learning the client's input, ensuring the privacy of the client's data remains protected. The system gurantees that all honest clients' inputs are included in final result and all malformed inputs from malicious clients are excluded from final result. Even though a minority of malicious servers exist, they will not prevent the output party computing the final result.

## Preparing Configuration and Input 
Each party has a configuration file comprising parameters, alongside an input file. The client's input file has response to the experiment or survey. The server and output party input files have experiment details like the experiment due time. Scripts to generate configuration and input files are located at each party's scripts/generator. To ensure each party runs successfully, configuration and input files must be prepared in advance.

## Running Computation
There are three ways running computation:
1. Seperate execution. This allows each party running on a different local machine. 
   
   To run a server with TLS (default), at the folder server/cmd, compile then run
   ```
   ./cmd -confpath="path_to_server_config_file" -inputpath="path_to_experiments_file" -logpath="path_to_log_folder" -n_client=num_of_clients
   ```
   Note: use -mode="http" to run without TLS, default setting is using TLS; use -n_client_mal=num_of_malicious_client to run experiment with malicious clients, default setting is assuming clients are all honest.

   To run an output party with TLS (default), at the folder outputparty/cmd, compile then run
   ```
   ./cmd -confpath="path_to_output_party_config_file" -inputpath="path_to_experiments_file" -logpath="path_to_log_folder" -n_client=num_of_clients
   ```
   Note: use -mode="http" to run without TLS, default setting is using TLS

   To run a client, at the folder client/cmd, compile then run
   ```
   ./cmd -confpath=“path_to_client_config_file” -inputpath=“path_to_input_file” -logpath="path_to_log_folder"
   ``` 
   Note: use -mode=honest to run client without malicious behaviour, default setting is malicious client
   
   **Note:** Servers and output party need to start running before clients.
2. One-command local execution. This allows all parties running on the same machine.
   At folder local, after preparing each party's template, compile then run
   ```
   ./local
   ``` 
3. Cloud deployment and execution. This allows all parties running on cloud machines. Detailes are described in the following repository.
https://github.com/GUSecLab/smc-in-a-box-ansible/tree/main







  
