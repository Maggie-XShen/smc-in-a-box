# SCIF: Privacy-Preserving Statistics Collection with Input Validation and Full Security
This repository contains the prototype implementation of SCIF, appearing in Proceedings on Privacy Enhancing Technologies (PoPETS), 2025.

SCIF is a practical system for private statistics,  comprising clients, several secure multiparty computation servers, and output party(e.g., data analyst) written in Golang. The system guarantees that all honest clients' inputs are included in final result and all malformed inputs from malicious clients are excluded from final result. Even though a minority of malicious servers exist, they will not prevent the output party computing the final result. For more details, please refer to the [paper](https://eprint.iacr.org/2024/1821).

## Build & Run with Docker Compose
One-command local execution. This allows all parties running on the same machine.
At folder local, after preparing each party's template, compile then run
```
./local
```

## Build & Run on Cloud Provider
The deployment and execution on Google Cloud are managed through ansible. Details are described in the following repository.
https://github.com/GUSecLab/smc-in-a-box-ansible/tree/main

## Build & Run Manually
### 1. Building environment
   
   First, ensure that you have Go installed. 

   ```
   % go version
   ```
   
   Second, ensure that MySQL is installed and running with a configured username and password for connections. The default credentials in SetupDatabase() (store.go) may need to be replaced with your own.

   ```
   % mysql --version 
   ```

   If MySQL is installed, you’ll see a version number.

   ```
   % brew services list
   ```

   If MySQL is running, it will be marked as started.
   
### 2. Prepare config file and input file
   
Each client, server and output party should obtain a config file and an input file before the software runs. The config file is a JSON blob, and a template of it can be found at client_template.json, server_template.json and outputparty_template.json. Scripts to generate config files and input files for many clients, servers and output parties are located at each party's scripts/generator.

Example of the config file for a client is shown below. 
```
{
    "Client_ID": "c1",
    "Token": "tk1",
    "URLs": ["http://127.0.0.1:50001/client/", 
            "http://127.0.0.1:50002/client/", 
            "http://127.0.0.1:50003/client/", 
            "http://127.0.0.1:50004/client/"], 
    "N": 4,  
    "T": 1,  
    "Q":  41543, 
    "N_secrets":10000, 
    "M": 50, 
    "N_open":240 
}
```

To describe the meaning of important fields:
- URLs: a list of server URLs for clients submitting their data to each server.
- N: number of total servers
- T: number of malicious servers
- Q: a modulus
- N_secrets: length of a client's input vector
- M: row numbers of extended witness in ligero ZK proof
- N_open: number of opened columns in encoded extended witness

Example of the config file for a server is shown below.
```
{
    "Server_ID": "s1",
    "Token": "stk1",
    "Cert_path":"/path to certificate", 
    "Key_path":"/path to private key",  
    "Port": "50001", 
    "Complaint_urls":[
        "http://127.0.0.1:50002/complaint/", 
        "http://127.0.0.1:50003/complaint/", 
        "http://127.0.0.1:50004/complaint/"  
    ],
    "Masked_share_urls":[
        "http://127.0.0.1:50002/maskedShare/", 
        "http://127.0.0.1:50003/maskedShare/", 
        "http://127.0.0.1:50004/maskedShare/"  
    ],
    "Share_Index": 1, 
    "N": 4, 
    "T": 1, 
    "Q":  41543, 
    "N_secrets":10000,
    "M": 50,  
    "N_open":240
}
```

To describe the meaning of important fields:

- Cert_path: server's certificate location required when running with cloud provider.
- Key_path: server's private key location required when running with cloud provider.
- Port: the port on which clients and other servers will connect to server.
- Complaint_urls: a list of server URLs for a server submitting complaints to other servers.
- Masked_share_urls: a list of server URLs for a server submitting masked shares to other servers.
- Share_Index: index number associated to server ID, e.g. 1 for server s1, 2 for server s2


Examples of the input files for client, server and output party are shown below.

client
```
[
   {"Exp_ID":"exp1","Secrets":[0,1]},
   {"Exp_ID":"exp2","Secrets":[1,1]}
]
```

server
```
[
   {"Exp_ID":"exp1",
   "ClientShareDue":"2025-03-11 18:13:57.188395 +0000 UTC",  
   "ComplaintDue":"2025-03-11 18:15:57.188395 +0000 UTC", 
   "ShareBroadcastDue":"2025-03-11 18:17:57.188395 +0000 UTC", 
   "Owner":"http://127.0.0.1:60000/serverShare/"}
]
```

output party
```
[
   {"Exp_ID":"exp1",
   "ClientShareDue":"2025-03-11 18:13:57.188395 +0000 UTC",
   "ServerShareDue":"2025-03-11 18:19:57.188395 +0000 UTC"} 
]
```

To describe the meaning of important fields:

- Secretes: a vector of a client's input, each bit represents an attribute.
- ClientShareDue: due time that clients submit shares and proofs to servers.
- ComplaintDue:: due time that servers submit their complaints to each other.
- ShareBroadcastDue: due time that servers submit masked shares to each other.
- ServerShareDue: due time that servers submit aggregated shares to the output party.
- Owner: a URL of the output party for servers submitting aggregated shares to the output party.

### 3. Run the software
   
To start up a server, at the folder server/cmd, compile then run
```
./cmd -confpath="path_to_server_config_file" -inputpath="path_to_experiments_file" -logpath="path_to_log_folder" -n_client=num_of_clients
```

To start up an output party, at the folder outputparty/cmd, compile then run
```
./cmd -confpath="path_to_output_party_config_file" -inputpath="path_to_experiments_file" -logpath="path_to_log_folder" -n_client=num_of_clients
```

To start up a client, at the folder client/cmd, compile then run
```
./cmd -confpath=“path_to_client_config_file” -inputpath=“path_to_input_file” -logpath="path_to_log_folder"
``` 

To describe the meaning of some parameters:
- For server and output party, use -mode="http" to disable TLS; the default enables it (which requires setup of certificate).
- For server, use -n_client_mal=num_of_malicious_client to enable malicious clients; the default assumes all are honest.
- For client, use -mode=honest to run client without malicious behavior. Default setting assumes client could act maliciously.
   
 **Note:** Servers and the output party must start before clients.


## Citation
If you find this work useful, please cite it as follows:




  
